<script lang="ts">
  import { goto } from '$app/navigation';
  import { login } from '$lib/stores/auth.svelte';

  let username = $state('');
  let password = $state('');
  let loading = $state(false);
  let error = $state('');

  let isValid = $derived(username.trim() !== '' && password.trim() !== '');

  async function handleSubmit(e: SubmitEvent) {
    e.preventDefault();
    if (!isValid || loading) return;

    error = '';
    loading = true;

    try {
      const ok = await login(username, password);
      if (ok) {
        await goto('/');
      } else {
        error = 'Invalid username or password.';
      }
    } catch (_err) {
      error = 'An error occurred. Please try again.';
    } finally {
      loading = false;
    }
  }
</script>

<svelte:head>
  <title>Login — FreeMED EMR</title>
</svelte:head>

<div class="flex items-center justify-center min-h-[calc(100vh-6rem)]">
  <div class="w-full max-w-sm">
    <div class="bg-white rounded-xl shadow-lg p-8">
      <div class="text-center mb-6">
        <h1 class="text-2xl font-bold text-gray-800">FreeMED EMR</h1>
        <p class="text-gray-500 mt-1 text-sm">Sign in to your account</p>
      </div>

      {#if error}
        <div
          class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg mb-6 text-sm"
          role="alert"
        >
          {error}
        </div>
      {/if}

      <form onsubmit={handleSubmit} class="space-y-5">
        <div>
          <label for="username" class="block text-sm font-medium text-gray-700 mb-1">
            Username
          </label>
          <input
            id="username"
            type="text"
            bind:value={username}
            required
            autocomplete="username"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm
                   placeholder-gray-400 text-sm
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                   disabled:bg-gray-100 disabled:cursor-not-allowed"
            placeholder="Enter your username"
            disabled={loading}
          />
        </div>

        <div>
          <label for="password" class="block text-sm font-medium text-gray-700 mb-1">
            Password
          </label>
          <input
            id="password"
            type="password"
            bind:value={password}
            required
            autocomplete="current-password"
            class="w-full px-3 py-2 border border-gray-300 rounded-lg shadow-sm
                   placeholder-gray-400 text-sm
                   focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500
                   disabled:bg-gray-100 disabled:cursor-not-allowed"
            placeholder="Enter your password"
            disabled={loading}
          />
        </div>

        <button
          type="submit"
          disabled={!isValid || loading}
          class="w-full flex items-center justify-center px-4 py-2.5 border border-transparent
                 rounded-lg shadow-sm text-sm font-medium text-white
                 bg-blue-600 hover:bg-blue-700
                 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500
                 disabled:opacity-50 disabled:cursor-not-allowed
                 transition-colors"
        >
          {#if loading}
            <svg
              class="animate-spin -ml-1 mr-2 h-4 w-4 text-white"
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
            >
              <circle
                class="opacity-25"
                cx="12"
                cy="12"
                r="10"
                stroke="currentColor"
                stroke-width="4"
              />
              <path
                class="opacity-75"
                fill="currentColor"
                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
              />
            </svg>
            Signing in…
          {:else}
            Sign in
          {/if}
        </button>
      </form>
    </div>
  </div>
</div>
