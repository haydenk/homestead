"use strict";

// ── Init ─────────────────────────────────────────────────────────────────────
document.addEventListener("DOMContentLoaded", () => {
  bindEvents();
  connectWS();
});

// ── WebSocket ─────────────────────────────────────────────────────────────────
let _ws = null;
let _wsReconnectTimer = null;

function connectWS() {
  clearTimeout(_wsReconnectTimer);
  const proto = location.protocol === "https:" ? "wss:" : "ws:";
  const ws = (_ws = new WebSocket(`${proto}//${location.host}/ws`));

  ws.onmessage = (e) => {
    try {
      updateStatusDots(JSON.parse(e.data));
    } catch (_) {}
  };

  ws.onclose = () => {
    if (ws === _ws) _wsReconnectTimer = setTimeout(connectWS, 3000);
  };

  ws.onerror = () => ws.close();
}

function updateStatusDots(statuses) {
  document.querySelectorAll(".status-dot[data-id]").forEach((dot) => {
    const s = statuses[dot.dataset.id];
    if (!s) return;
    dot.classList.remove("up", "slow", "down", "unknown");
    const cls = s.up ? (s.responseTimeMs > 1000 ? "slow" : "up") : "down";
    dot.classList.add(cls);
    dot.dataset.tooltip = s.up ? `Online \u00b7 ${s.responseTimeMs}ms` : `Offline${s.error ? `: ${s.error}` : ""}`;
    dot.setAttribute("aria-label", `status ${cls}`);
  });
}

// ── Search / filter ───────────────────────────────────────────────────────────
function applySearch(raw) {
  const q = raw.trim().toLowerCase();
  let totalVisible = 0;

  document.querySelectorAll(".section").forEach((sec) => {
    let visible = 0;
    sec.querySelectorAll(".card").forEach((card) => {
      const match = !q || card.dataset.title.includes(q) || card.dataset.desc.includes(q) || card.dataset.tags.includes(q);
      card.hidden = !match;
      if (match) visible++;
    });
    sec.hidden = visible === 0;
    totalVisible += visible;
  });

  const emptyState = document.getElementById("empty-state");
  emptyState.hidden = totalVisible > 0 || !q;
  if (q) document.getElementById("empty-query").textContent = q;
}

// ── Refresh ───────────────────────────────────────────────────────────────────
async function triggerRefresh() {
  const btn = document.getElementById("btn-refresh");
  btn.classList.add("spinning");
  try {
    await fetch("/api/reload", { method: "POST" });
    window.location.reload();
  } catch {
    btn.classList.remove("spinning");
  }
}

// ── Event bindings ────────────────────────────────────────────────────────────
function bindEvents() {
  const search = document.getElementById("search");
  search.addEventListener("input", (e) => applySearch(e.target.value));
  document.getElementById("btn-refresh").addEventListener("click", triggerRefresh);

  document.addEventListener("keydown", (e) => {
    if ((e.key === "/" || (e.ctrlKey && e.key === "k")) && document.activeElement !== search) {
      e.preventDefault();
      search.focus();
      search.select();
      return;
    }
    if (e.key === "Escape" && document.activeElement === search) {
      search.value = "";
      applySearch("");
      search.blur();
      return;
    }
    if (e.key === "r" && !isTyping()) {
      triggerRefresh();
    }
  });
}

// ── Helpers ───────────────────────────────────────────────────────────────────
function isTyping() {
  const tag = document.activeElement?.tagName;
  return tag === "INPUT" || tag === "TEXTAREA" || document.activeElement?.isContentEditable;
}
