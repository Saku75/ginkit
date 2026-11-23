<script lang="ts">
  import type { PageProps } from "./$types";

  let { data }: PageProps = $props();
</script>

<main class="mx-auto max-w-xl p-8 font-sans">
  <h1 class="text-2xl font-semibold">ginkit</h1>
  <p class="mt-4">The API says: {data.message}</p>
  <p class="mt-4"><a class="underline" href="/api">API reference</a></p>
</main>
