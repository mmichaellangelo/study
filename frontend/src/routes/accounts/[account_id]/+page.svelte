<script lang="ts">
    import { onMount } from 'svelte';

    let { data } = $props()
    let loading = $state(true)

    onMount(() => {
        loading = false
    })
</script>
<h2>account</h2>
{#if loading}
    <p>loading...</p>
{:else}
    {#if data.account}
        <p>{data.account.username}</p>
        <p>{data.account.email}</p>
        <p>created {data.account.created?.toLocaleString().toLowerCase()}</p>
    {:else}
        <p>error getting account</p>
    {/if}
{/if}