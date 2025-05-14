<script lang="ts">
    import { onMount } from 'svelte';

    let { data } = $props()
    let loading = $state(true)

    onMount(() => {
        loading = false
    })
</script>
<a href="/study">back</a>
<h2>account</h2>
<div id="account_body">
    {#if loading}
        <p>loading...</p>
    {:else}
        {#if data.account}
            <p>username: {data.account.username}</p>
            <p>email: {data.account.email}</p>
            <p>created: {data.account.created?.toLocaleString().toLowerCase()}</p>
        {:else}
            <p>error getting account</p>
        {/if}
    {/if}
</div>

<style>
    #account_body {
        padding: 1rem;
        padding-top: 0rem;
    }
</style>