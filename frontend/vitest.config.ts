import { defineConfig, mergeConfig } from 'vitest/config';
import { fileURLToPath, URL } from 'node:url';
import vue from '@vitejs/plugin-vue';

export default mergeConfig(
  // Inline the minimal vite setup needed — avoids pulling in proxy/server config.
  {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
  },
  defineConfig({
    test: {
      environment: 'happy-dom',
      globals: true,
      setupFiles: ['./src/test/setup.ts'],
      include: ['src/**/*.test.ts'],
      coverage: {
        provider: 'v8',
        reporter: ['text', 'lcov'],
        include: ['src/**/*.ts'],
        exclude: ['src/api/generated/**', 'src/test/**'],
      },
    },
  }),
);
