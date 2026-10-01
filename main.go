// mxs looks up Minecraft player IDs: Bedrock XUID, Java UUID, Floodgate UUID.
package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const usage = `usage:
  mxs bedrock <gamertag>                             Bedrock player info
  mxs bedrock -r <xuid | floodgate uuid>             Bedrock player info
  mxs java <account name>                            Java player info
  mxs java -r <uuid>                                 Java player info`

var client = &http.Client{Timeout: 10 * time.Second}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, usage)
		os.Exit(2)
	}
	fs := flag.NewFlagSet(os.Args[1], flag.ExitOnError)
	fs.Usage = func() { fmt.Fprintln(os.Stderr, usage) }
	reverse := fs.Bool("r", false, "reverse lookup")
	_ = fs.Parse(os.Args[2:]) // ExitOnError: Parse exits instead of returning an error
	if fs.NArg() != 1 {
		fs.Usage()
		os.Exit(2)
	}
	out, err := run(os.Args[1], fs.Arg(0), *reverse)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mxs:", err)
		os.Exit(1)
	}
	fmt.Println(out)
}

func run(cmd, arg string, reverse bool) (string, error) {
	switch {
	case cmd == "bedrock" && reverse:
		xuid, err := parseXUID(arg)
		if err != nil {
			return "", err
		}
		gamertag, err := bedrockGamertag(xuid)
		if err != nil {
			return "", err
		}
		return bedrockInfo(gamertag, xuid), nil
	case cmd == "bedrock":
		xuid, err := bedrockXUID(arg)
		if err != nil {
			return "", err
		}
		return bedrockInfo(arg, xuid), nil
	case cmd == "java" && reverse:
		h := strings.ReplaceAll(arg, "-", "")
		if _, err := hex.DecodeString(h); err != nil || len(h) != 32 {
			return "", fmt.Errorf("invalid uuid %q", arg)
		}
		return javaProfile("https://sessionserver.mojang.com/session/minecraft/profile/"+h, "java uuid "+arg)
	case cmd == "java":
		return javaProfile("https://api.mojang.com/users/profiles/minecraft/"+url.PathEscape(arg), fmt.Sprintf("java player %q", arg))
	default:
		return "", fmt.Errorf("unknown command %q\n%s", cmd, usage)
	}
}

func bedrockInfo(gamertag string, xuid uint64) string {
	return fmt.Sprintf("Gamertag: %s\nXUID(DEC): %d\nXUID(HEX): %x\nFloodgate UUID: %s", gamertag, xuid, xuid, floodgateUUID(xuid))
}

// parseXUID accepts a decimal XUID or a Floodgate UUID.
func parseXUID(s string) (uint64, error) {
	if strings.Contains(s, "-") {
		return parseFloodgateUUID(s)
	}
	xuid, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid xuid %q", s)
	}
	return xuid, nil
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

func bedrockGamertag(xuid uint64) (string, error) {
	var res struct {
		Gamertag string `json:"gamertag"`
	}
	if err := getJSON("https://api.geysermc.org/v2/xbox/gamertag/"+strconv.FormatUint(xuid, 10), &res); err != nil {
		return "", fmt.Errorf("bedrock xuid %d: %w", xuid, err)
	}
	if res.Gamertag == "" {
		return "", fmt.Errorf("bedrock xuid %d not found", xuid)
	}
	return res.Gamertag, nil
}

// javaProfile fetches a Mojang profile; both endpoints return {"id", "name"}.
func javaProfile(u, what string) (string, error) {
	var res struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := getJSON(u, &res); err != nil {
		return "", fmt.Errorf("%s: %w", what, err)
	}
	if len(res.ID) != 32 || res.Name == "" {
		return "", fmt.Errorf("%s not found", what)
	}
	return fmt.Sprintf("Name: %s\nUUID: %s", res.Name, dashed(res.ID)), nil
}

// floodgateUUID matches Floodgate's `new UUID(0, xuid)`.
func floodgateUUID(xuid uint64) string {
	return dashed(fmt.Sprintf("%016x%016x", 0, xuid))
}

// parseFloodgateUUID is the inverse of floodgateUUID.
func parseFloodgateUUID(uuid string) (uint64, error) {
	h := strings.ReplaceAll(uuid, "-", "")
	if len(h) != 32 || h[:16] != strings.Repeat("0", 16) {
		return 0, fmt.Errorf("invalid floodgate uuid %q", uuid)
	}
	xuid, err := strconv.ParseUint(h[16:], 16, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid floodgate uuid %q", uuid)
	}
	return xuid, nil
}

func dashed(h string) string {
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func getJSON(u string, v any) error {
	resp, err := client.Get(u)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }() // read-only body: close error carries nothing actionable
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
