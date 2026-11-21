<script lang="ts">
  import { page } from "$app/state";
</script>

<main class="mx-auto max-w-xl p-8 font-sans">
  <h1 class="text-2xl font-semibold">{page.status}</h1>
  <p class="mt-4">{page.error?.message}</p>
  <p class="mt-4"><a class="underline" href="/">Back to the start</a></p>
</main>
