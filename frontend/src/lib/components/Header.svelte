<script>
    import { userState } from "$lib/state/account.svelte";
    import { onMount } from "svelte";

    let isLoading = $state(true)

    onMount(() => {
        isLoading = false;
    })

</script>

<div id="header">
    <div id="title_user_container">
        <a href="/">
            <h1>disco</h1>
        </a>
        <div id="user_container">
            {#if isLoading}
                <p>loading...</p>
            {:else}
                {#if userState.ID != -1}
                    <p>logged in as <a href={`/accounts/${userState.ID}`}>{userState.Username}</a></p>
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
</div>



<style>
    #header {
        background-color: var(--col-purplegrey);
        padding: 1rem;
    }

    #title_user_container {
        display: flex;
        align-items: center;
    }

    #user_container {
        margin-left: auto;
        text-align: right;
    }

    h1 {
        color: var(--col-lightpink);
    }
</style>



