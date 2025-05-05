<script>
    import { userState } from "$lib/state/account.svelte";

</script>
<h1>disco</h1>
<nav>
    <a href="/study">study</a>
    <a href="/create">create</a>
    <a href="/explore">explore</a>
    {#if userState.ID != -1}
        <p>Logged in as {userState.ID}</p>
    {:else}
        <a href="/login">login</a>
        <a href="/register">register</a>
    {/if}
</nav>