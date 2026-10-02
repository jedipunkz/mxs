// Mirrors ../../main_test.go. Run: npm test
import assert from "node:assert/strict";
import { test } from "node:test";
import { floodgateUUID, javaUUID, parseFloodgateUUID, parseXUID, run } from "./mxs.ts";

test("uuid format", () => {
  assert.equal(floodgateUUID(2535428654845283n), "00000000-0000-0000-0009-01f57c534d63");
  assert.equal(parseFloodgateUUID("00000000-0000-0000-0009-01f57c534d63"), 2535428654845283n);
  assert.throws(() => parseFloodgateUUID("069a79f4-44e9-4726-a5be-fca90e38aaf5"));
  for (const s of ["2535414915229641", "901f2496167c9", "901F2496167C9", "0x901f2496167c9", "00000000-0000-0000-0009-01f2496167c9", "0000000000000000000901f2496167c9"]) {
    assert.equal(parseXUID(s), 2535414915229641n, s);
  }
  assert.throws(() => parseXUID("not-an-id"));
  assert.throws(() => parseXUID("18446744073709551616ff")); // > uint64
  for (const [s, want] of [
    ["069a79f4-44e9-4726-a5be-fca90e38aaf5", true],
    ["069A79F444E94726A5BEFCA90E38AAF5", true],
    ["00000000-0000-0000-0009-01f2496167c9", false], // floodgate
    ["901f2496167c9", false], // hex xuid
    ["2535414915229641", false], // dec xuid
  ] as const) {
    assert.equal(javaUUID(s) !== undefined, want, s);
  }
});

const stub = (status: number, body: string) => async () => new Response(status === 204 ? null : body, { status });

test("run", async () => {
  // XUID above 2^53 keeps full precision.
  assert.deepEqual(await run("bedrock", "Big", stub(200, '{"xuid":18446744073709551615}')), [
    ["Gamertag", "Big"],
    ["XUID(DEC)", "18446744073709551615"],
    ["XUID(HEX)", "ffffffffffffffff"],
    ["Floodgate UUID", "00000000-0000-0000-ffff-ffffffffffff"],
  ]);
  assert.deepEqual(await run("reverse", "069a79f4-44e9-4726-a5be-fca90e38aaf5", stub(200, '{"id":"069a79f444e94726a5befca90e38aaf5","name":"Notch"}')), [
    ["Name", "Notch"],
    ["UUID", "069a79f4-44e9-4726-a5be-fca90e38aaf5"],
  ]);
  await assert.rejects(run("java", "nobody", stub(404, "")), { message: 'java player "nobody" not found' });
  await assert.rejects(run("java", "nobody", stub(204, "")), { message: 'java player "nobody" not found' });
  await assert.rejects(run("bedrock", "x", stub(503, '{"message":"Unable to find user in our cache"}')), {
    message: 'bedrock player "x": HTTP 503 Unable to find user in our cache',
  });
});
