<script>
    import { userState } from "$lib/state/account.svelte";
    import { onMount } from "svelte";

    let isLoading = $state(true)

    onMount(() => {
        isLoading = false;
    })

</script>

<h1>disco</h1>
<nav>
    <a href="/study">study</a>
    <a href="/create">create</a>
    <a href="/explore">explore</a>
</nav>

{#if isLoading}
    <p>loading...</p>
{:else}
    {#if userState.ID != -1}
        <p>Logged in as {userState.Username}</p>
        <a href="/logout">log out</a>
    {:else}
    <a href="/login">login</a>
    <a href="/register">register</a>
    {/if}
{/if}

