import { defineConfig } from 'astro/config';
import vue from '@astrojs/vue';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  site: 'https://felixgeelhaar.github.io',
  base: '/acai',
  integrations: [vue()],
  vite: { plugins: [tailwindcss()] },
  output: 'static',
});
