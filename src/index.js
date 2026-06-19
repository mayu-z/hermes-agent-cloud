import OpenAI from "openai";

const NVIDIA_BASE_URL = "https://integrate.api.nvidia.com/v1";
const NEMOTRON_MODEL  = "nvidia/nemotron-3-ultra-550b-a55b";

const HERMES_SYSTEM = `You are Hermes, an intelligent AI agent.
You are helpful, concise, and honest.
You remember the conversation history within a session.
When asked to reason through a problem, think step-by-step.`;

// ── KV helpers ───────────────────────────────────────────────────────────────

async function loadMemory(kv, sessionId) {
  const raw = await kv.get(`session:${sessionId}`);
  return raw ? JSON.parse(raw) : { messages: [] };
}

async function saveMemory(kv, sessionId, memory) {
  await kv.put(`session:${sessionId}`, JSON.stringify(memory), {
    expirationTtl: 86400, // auto-expire after 24h
  });
}

// ── Route handlers ───────────────────────────────────────────────────────────

async function handleChat(request, env) {
  let body;
  try { body = await request.json(); }
  catch { return jsonError("Invalid JSON", 400); }

  const { message, sessionId = "default" } = body;
  if (!message) return jsonError("Missing 'message' field", 400);

  const memory = await loadMemory(env.HERMES_MEMORY, sessionId);
  memory.messages.push({ role: "user", content: message });

  const client = new OpenAI({
    apiKey: env.NVIDIA_API_KEY,
    baseURL: NVIDIA_BASE_URL,
  });

  let reply;
  try {
    const res = await client.chat.completions.create({
      model: NEMOTRON_MODEL,
      messages: [{ role: "system", content: HERMES_SYSTEM }, ...memory.messages],
      max_tokens: 1024,
      temperature: 0.7,
    });
    reply = res.choices[0].message.content;
  } catch (err) {
    return jsonError(`NVIDIA API error: ${err.message}`, 502);
  }

  memory.messages.push({ role: "assistant", content: reply });
  await saveMemory(env.HERMES_MEMORY, sessionId, memory);

  return json({ reply, sessionId, turns: memory.messages.length / 2 });
}

async function handleClear(request, env) {
  const body = await request.json().catch(() => ({}));
  const { sessionId = "default" } = body;
  await env.HERMES_MEMORY.delete(`session:${sessionId}`);
  return json({ cleared: true, sessionId });
}

// ── Main ─────────────────────────────────────────────────────────────────────

export default {
  async fetch(request, env) {
    const { pathname } = new URL(request.url);
    const method = request.method;

    if (pathname === "/")                               return json({ status: "Hermes online ✅" });
    if (pathname === "/chat"  && method === "POST")     return handleChat(request, env);
    if (pathname === "/clear" && method === "POST")     return handleClear(request, env);

    return new Response("Not found", { status: 404 });
  },
};

// ── Utils ────────────────────────────────────────────────────────────────────

const json      = (data, status = 200) => new Response(JSON.stringify(data), { status, headers: { "Content-Type": "application/json" } });
const jsonError = (msg,  status)       => json({ error: msg }, status);