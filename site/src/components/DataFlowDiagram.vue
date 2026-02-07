<template>
  <div class="flow-diagram">
    <!-- Tabs -->
    <div class="flex gap-1 mb-6 bg-[var(--color-bg-card)] rounded-lg p-1 border border-[var(--color-border)]">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="activeTab = tab.id"
        :class="[
          'flex-1 px-4 py-2 rounded-md text-sm font-medium transition-all duration-200',
          activeTab === tab.id
            ? `bg-[var(--color-bg-primary)] shadow-sm`
            : 'text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
        ]"
        :style="activeTab === tab.id ? { color: tab.color } : {}"
      >
        {{ tab.label }}
      </button>
    </div>

    <!-- Diagram -->
    <div class="relative bg-[var(--color-bg-card)] rounded-lg border border-[var(--color-border)] p-6 overflow-hidden">
      <!-- Read Path -->
      <div v-if="activeTab === 'read'" class="flow-nodes">
        <div class="flow-node" style="--node-color: var(--color-accent)">
          <div class="flow-node-icon">☁️</div>
          <div class="flow-node-label">Granola Cloud API</div>
          <div class="flow-node-detail">External data source</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-accent)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-accent)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-accent)">
          <div class="flow-node-icon">🔌</div>
          <div class="flow-node-label">Granola Client</div>
          <div class="flow-node-detail">Anti-corruption layer: API DTOs → domain types</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-red)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-red)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-red)">
          <div class="flow-node-icon">🔄</div>
          <div class="flow-node-label">Resilient Repo</div>
          <div class="flow-node-detail">Circuit breaker, retry, rate limit, timeout</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-cyan)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-cyan)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-cyan)">
          <div class="flow-node-icon">💾</div>
          <div class="flow-node-label">Cached Repo</div>
          <div class="flow-node-detail">SQLite local cache (TTL 15min)</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-accent-secondary)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-accent-secondary)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-accent-secondary)">
          <div class="flow-node-icon">⚡</div>
          <div class="flow-node-label">Use Cases → MCP Server / CLI</div>
          <div class="flow-node-detail">Application layer → Interface adapters</div>
        </div>
      </div>

      <!-- Write Path -->
      <div v-if="activeTab === 'write'" class="flow-nodes">
        <div class="flow-node" style="--node-color: var(--color-green)">
          <div class="flow-node-icon">🤖</div>
          <div class="flow-node-label">Agent calls add_note / complete_action_item</div>
          <div class="flow-node-detail">Write request from MCP tool</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-green)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-green)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-green)">
          <div class="flow-node-icon">✅</div>
          <div class="flow-node-label">Use Case</div>
          <div class="flow-node-detail">Validates input, creates domain entity</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-green)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-green)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-cyan)">
          <div class="flow-node-icon">💾</div>
          <div class="flow-node-label">Local SQLite Store</div>
          <div class="flow-node-detail">Persists to local.db (separate from cache.db)</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-accent-secondary)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-accent-secondary)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-accent-secondary)">
          <div class="flow-node-icon">📤</div>
          <div class="flow-node-label">Outbox Dispatcher</div>
          <div class="flow-node-detail">Persists event for future upstream sync</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-accent)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-accent)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-accent)">
          <div class="flow-node-icon">📡</div>
          <div class="flow-node-label">Event Dispatcher → MCP Notifier</div>
          <div class="flow-node-detail">Real-time updates to subscribed sessions</div>
        </div>
      </div>

      <!-- Policy Path -->
      <div v-if="activeTab === 'policy'" class="flow-nodes">
        <div class="flow-node" style="--node-color: var(--color-orange)">
          <div class="flow-node-icon">🤖</div>
          <div class="flow-node-label">Agent calls get_transcript</div>
          <div class="flow-node-detail">Request with tags: ["confidential"]</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-orange)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-orange)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-orange)">
          <div class="flow-node-icon">🛡️</div>
          <div class="flow-node-label">Policy Middleware</div>
          <div class="flow-node-detail">Extract meeting context, evaluate ACL rules</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-orange)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-orange)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-red)">
          <div class="flow-node-icon">🚫</div>
          <div class="flow-node-label">ACL Check</div>
          <div class="flow-node-detail">First-match-wins — deny returns error immediately</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-green)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-green)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-green)">
          <div class="flow-node-icon">✂️</div>
          <div class="flow-node-label">Redaction Engine</div>
          <div class="flow-node-detail">Emails → [EMAIL], speakers → Speaker N, keywords → [REDACTED]</div>
        </div>
        <div class="flow-connector" style="--conn-color: var(--color-accent)">
          <svg viewBox="0 0 24 40" class="w-6 h-10 mx-auto"><path d="M12 0 L12 30 M6 24 L12 30 L18 24" stroke="currentColor" stroke-width="2" fill="none" class="animate-dash" style="color: var(--color-accent)"/></svg>
        </div>
        <div class="flow-node" style="--node-color: var(--color-accent)">
          <div class="flow-node-icon">✅</div>
          <div class="flow-node-label">Redacted Response</div>
          <div class="flow-node-detail">Clean JSON returned to agent</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const activeTab = ref('read');

const tabs = [
  { id: 'read', label: 'Read Path', color: 'var(--color-accent)' },
  { id: 'write', label: 'Write Path', color: 'var(--color-green)' },
  { id: 'policy', label: 'Policy Enforcement', color: 'var(--color-orange)' },
];
</script>

<style scoped>
.flow-nodes {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0;
}

.flow-node {
  width: 100%;
  max-width: 480px;
  padding: 1rem 1.25rem;
  border-radius: 0.5rem;
  border: 1px solid color-mix(in srgb, var(--node-color) 30%, transparent);
  background: color-mix(in srgb, var(--node-color) 5%, var(--color-bg-primary));
  transition: all 0.2s ease;
}

.flow-node:hover {
  border-color: color-mix(in srgb, var(--node-color) 50%, transparent);
  background: color-mix(in srgb, var(--node-color) 10%, var(--color-bg-primary));
}

.flow-node-icon {
  font-size: 1.25rem;
  margin-bottom: 0.25rem;
}

.flow-node-label {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--color-text-primary);
}

.flow-node-detail {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.flow-connector {
  padding: 0.125rem 0;
}

@keyframes dash-flow {
  to {
    stroke-dashoffset: -20;
  }
}

.animate-dash {
  stroke-dasharray: 6 4;
  animation: dash-flow 1s linear infinite;
}
</style>
