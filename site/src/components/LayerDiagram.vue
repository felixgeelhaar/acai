<template>
  <div class="layer-diagram">
    <!-- Interactive layers -->
    <div class="layers-container">
      <div
        v-for="layer in layers"
        :key="layer.id"
        @mouseenter="activeLayer = layer.id"
        @mouseleave="activeLayer = null"
        :class="[
          'layer-ring',
          { 'layer-active': activeLayer === layer.id }
        ]"
        :style="{
          '--layer-color': layer.color,
          '--layer-size': layer.size,
        }"
      >
        <div class="layer-label">
          <span class="layer-name">{{ layer.name }}</span>
          <span class="layer-desc">{{ layer.desc }}</span>
        </div>
      </div>
    </div>

    <!-- Detail panel -->
    <div class="mt-6 rounded-lg border border-[var(--color-border)] bg-[var(--color-bg-card)] p-4 min-h-[120px] transition-all duration-200">
      <div v-if="activeLayer" class="detail-content">
        <h4 class="text-sm font-semibold mb-2" :style="{ color: activeDetail.color }">
          {{ activeDetail.name }}
        </h4>
        <p class="text-xs text-[var(--color-text-secondary)] mb-3">{{ activeDetail.description }}</p>
        <div class="flex flex-wrap gap-1.5">
          <span
            v-for="pkg in activeDetail.packages"
            :key="pkg"
            class="inline-block px-2 py-0.5 rounded text-xs font-mono"
            :style="{
              backgroundColor: `color-mix(in srgb, ${activeDetail.color} 10%, transparent)`,
              color: activeDetail.color,
              border: `1px solid color-mix(in srgb, ${activeDetail.color} 20%, transparent)`,
            }"
          >
            {{ pkg }}
          </span>
        </div>
      </div>
      <div v-else class="flex items-center justify-center h-full text-xs text-[var(--color-text-muted)]">
        Hover over a layer to explore
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

const activeLayer = ref<string | null>(null);

const layers = [
  { id: 'interfaces', name: 'Interfaces', desc: 'MCP + CLI', color: 'var(--color-accent)', size: '100%' },
  { id: 'application', name: 'Application', desc: 'Use Cases', color: 'var(--color-accent-secondary)', size: '78%' },
  { id: 'domain', name: 'Domain', desc: 'Pure Logic', color: 'var(--color-green)', size: '56%' },
  { id: 'infrastructure', name: 'Infrastructure', desc: 'Adapters', color: 'var(--color-orange)', size: '100%' },
];

const layerDetails: Record<string, { name: string; color: string; description: string; packages: string[] }> = {
  interfaces: {
    name: 'Interfaces Layer',
    color: 'var(--color-accent)',
    description: 'Inbound adapters that translate external requests into application use case calls. MCP server with typed tools/resources and CLI with Cobra commands.',
    packages: ['mcp/', 'cli/', 'PolicyMiddleware', 'MCPNotifier'],
  },
  application: {
    name: 'Application Layer',
    color: 'var(--color-accent-secondary)',
    description: 'One use case per file, each with Execute(ctx, input) (output, error). Orchestrates domain objects and repository calls.',
    packages: ['ListMeetings', 'GetTranscript', 'AddNote', 'ExportEmbeddings', 'SyncMeetings'],
  },
  domain: {
    name: 'Domain Layer',
    color: 'var(--color-green)',
    description: 'Pure Go, zero external dependencies. Business rules, aggregates, value objects, domain events, and repository port interfaces.',
    packages: ['Meeting', 'Transcript', 'ActionItem', 'AgentNote', 'Policy', 'DomainEvent'],
  },
  infrastructure: {
    name: 'Infrastructure Layer',
    color: 'var(--color-orange)',
    description: 'External adapters implementing domain ports. Granola API client, SQLite cache, resilience decorators, event dispatching.',
    packages: ['granola/', 'resilience/', 'cache/', 'localstore/', 'outbox/', 'webhook/'],
  },
};

const activeDetail = computed(() => {
  if (!activeLayer.value) return layerDetails.domain;
  return layerDetails[activeLayer.value];
});
</script>

<style scoped>
.layers-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.layer-ring {
  width: var(--layer-size);
  padding: 0.875rem 1.25rem;
  border-radius: 0.75rem;
  border: 1px solid color-mix(in srgb, var(--layer-color) 25%, transparent);
  background: color-mix(in srgb, var(--layer-color) 5%, var(--color-bg-primary));
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: center;
}

.layer-ring:hover,
.layer-active {
  border-color: color-mix(in srgb, var(--layer-color) 50%, transparent);
  background: color-mix(in srgb, var(--layer-color) 12%, var(--color-bg-primary));
  transform: scale(1.02);
}

.layer-label {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
}

.layer-name {
  font-weight: 600;
  font-size: 0.875rem;
  color: var(--layer-color);
}

.layer-desc {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}

.detail-content {
  animation: fadeIn 0.15s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
</style>
