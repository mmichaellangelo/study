<script>
    import { userState } from "$lib/state/account.svelte";
    import { onMount } from "svelte";

    let isLoading = $state(true)

    onMount(() => {
        isLoading = false;
    })

</script>

<div id="title_user_container">
    <h1>disco</h1>
    <div id="user_container">
        {#if isLoading}
            <p>loading...</p>
        {:else}
            {#if userState.ID != -1}
                <p>Logged in as <a href={`/accounts/${userState.ID}`}>{userState.Username}</a></p>
                <a href="/logout">log out</a>
            {:else}
                <a href="/login">login</a>
                <a href="/register">register</a>
            {/if}
        {/if}

    </div>
    
</div>

<nav>
    <a href="/study">study</a>
    <a href="/create">create</a>
    <a href="/explore">explore</a>
</nav>

<style>
    #title_user_container {
        display: flex;
        align-items: center;
    }

    #user_container {
        margin-left: auto;
        text-align: right;
    }
</style>



