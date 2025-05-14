<script lang="ts">
    import { API } from '$lib/api.js';
    import Loader from '$lib/components/Loader.svelte';
    import { GotoReload } from '$lib/functions/navigation.js';
    import { onMount } from 'svelte';

    let { data } = $props()
    let isLoading = $state(true)
    let success = $state(false)

    onMount(async () => {
        try {
        const response = await fetch(`${API}/auth/logout`, {
            method: "POST",
            credentials: "include",
        })
        if (!response.ok) {
            isLoading = false
            success = false
        }
        isLoading = false
        success = true
        setTimeout(() => { GotoReload("/") }, 1000)
    } catch (e) {
        isLoading = false
        success = false
    }
    })
</script>

{#if isLoading}
    <Loader />
{:else}
    {#if data}
        {#if success}
            <p>you're logged out!</p>
        {:else}
            <p>there was a problem logging you out. please try refreshing the page.</p>
        {/if}
    {/if}
{/if}

