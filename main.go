// mxs looks up Minecraft player IDs: Bedrock XUID, Java UUID, Floodgate UUID.
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const usage = `usage:
  mxs bedrock <gamertag>       Bedrock XUID
  mxs java <account name>      Java UUID
  mxs floodgate <gamertag>     Floodgate UUID`

var client = &http.Client{Timeout: 10 * time.Second}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	out, err := run(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "mxs:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

func run(cmd, name string) (string, error) {
	switch cmd {
	case "bedrock":
		xuid, err := bedrockXUID(name)
		return strconv.FormatUint(xuid, 10), err
	case "floodgate":
		xuid, err := bedrockXUID(name)
		return floodgateUUID(xuid), err
	case "java":
		return javaUUID(name)
	default:
		return "", fmt.Errorf("unknown command %q\n%s", cmd, usage)
	}
}

// ponytail: GeyserMC API only knows gamertags that have joined a Geyser server; full coverage needs Xbox Live auth.
func bedrockXUID(gamertag string) (uint64, error) {
	var res struct {
		XUID uint64 `json:"xuid"`
	}
	if err := getJSON("https://api.geysermc.org/v2/xbox/xuid/"+url.PathEscape(gamertag), &res); err != nil {
		return 0, fmt.Errorf("bedrock player %q: %w", gamertag, err)
	}
	if res.XUID == 0 {
		return 0, fmt.Errorf("bedrock player %q not found", gamertag)
	}
	return res.XUID, nil
}

func javaUUID(name string) (string, error) {
	var res struct {
		ID string `json:"id"`
	}
	if err := getJSON("https://api.mojang.com/users/profiles/minecraft/"+url.PathEscape(name), &res); err != nil {
		return "", fmt.Errorf("java player %q: %w", name, err)
	}
	if len(res.ID) != 32 {
		return "", fmt.Errorf("java player %q not found", name)
	}
	return dashed(res.ID), nil
}

// floodgateUUID matches Floodgate's `new UUID(0, xuid)`.
func floodgateUUID(xuid uint64) string {
	return dashed(fmt.Sprintf("%016x%016x", 0, xuid))
}

func dashed(h string) string {
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func getJSON(u string, v any) error {
	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Mojang answers not-found with 204/404 and an empty or error body; callers detect the zero value.
	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		// GeyserMC uses "message" (503 when the gamertag is not cached), Mojang uses "errorMessage".
		var e struct {
			Message      string `json:"message"`
			ErrorMessage string `json:"errorMessage"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&e) // best effort: the status code alone is still reported
		return fmt.Errorf("HTTP %d %s%s", resp.StatusCode, e.Message, e.ErrorMessage)
	}
	return json.NewDecoder(resp.Body).Decode(v)
}
