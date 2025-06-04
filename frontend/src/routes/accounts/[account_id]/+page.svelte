<svelte:head>
    <title>disco - account</title>
</svelte:head>
<script lang="ts">
    import { API } from '$lib/api';
    import Loader from '$lib/components/Loader.svelte';
    import { GotoReload } from '$lib/functions/navigation';
    import { userState } from '$lib/state/account.svelte';
    import { onMount } from 'svelte';

    let { data } = $props()
    let loading = $state(true)
    let password = $state("")
    let message = $state("")

    onMount(() => {
        loading = false
    })

    let dialogElement = $state<HTMLDialogElement>()
    let dialogPage = $state(0)
    let isDeleting = $state(false)

    function showDialog() {
        dialogPage = 0
        dialogElement?.showModal()
    }

    function closeDialog() {
        dialogElement?.close()
    }

    async function deleteAccount(e: SubmitEvent) {
        e.preventDefault()
        isDeleting = true 
        const credentials = btoa(`${data.account?.username}:${password}`)
        try {
            const res = await fetch(`${API}/accounts/${data.account?.id}`, {
                method: "DELETE",
                credentials: 'include',
                headers: {
                    Authorization: `Basic ${credentials}`
                }
            })
            if (!res.ok) {
                isDeleting = false
                message = "an error occurred"
                return
            }
            isDeleting = false
            message = "account deleted"
            userState.ID = -1
            userState.Username = ""
            setTimeout(() => GotoReload("/"), 1000)
        } catch (e) {
            console.log(e)
        }
    }
</script>

<dialog bind:this={dialogElement}>
    {#if dialogPage == 0}
        <p>are you sure you want to delete your account? this can't be undone!</p>
        <div>
            <button onclick={() => dialogPage = 1} class="delete">yes</button>
            <button onclick={closeDialog}>no</button>
        </div>
    {:else if dialogPage == 1}
        <form onsubmit={deleteAccount} id="delete_form">
            <label>
                please enter your password <br />
                <input name="password" type="password" bind:value={password}> <br />
                <button type="submit" class="delete" disabled={isDeleting}>delete account</button>
                <button onclick={closeDialog} disabled={isDeleting}>cancel</button>
                {#if isDeleting} <Loader /> {/if}
            </label>
        </form>
        <p>{message}</p>
    {/if}
</dialog>

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
            <button class="delete" onclick={showDialog}>delete account</button>
        {:else}
            <p>error getting account</p>
        {/if}
    {/if}
</div>



<style>
    #account_body {
        padding-top: 0rem;
    }

    dialog {
        position: absolute;
        background-color: var(--col-darkblue);
        backdrop-filter: blur(5px);
        color: var(--col-lightpink);
        border: 2px solid var(--col-msg-error);
        border-radius: 1rem;
        box-shadow: 0px 0px 10px var(--col-msg-error);
        font-weight: 600;
    }

    dialog>p {
        color: var(--col-msg-error);
    }

    dialog>div>button {
        margin-left: 1rem;
    }

    button:disabled {
        opacity: 0.5;
        cursor: default;
    }

    ::backdrop {
        backdrop-filter: blur(7px);
        background-color: rgba(0,0,0,0.6);
    }

    #delete_form {
        display: flex;
        flex-direction: column;
        align-items: left;
    }
</style>