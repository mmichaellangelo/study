<svelte:head>
    <title>disco - register</title>
</svelte:head>

<script lang="ts">
    import { API } from "$lib/api";
    import Loader from "$lib/components/Loader.svelte";
    import StatusMessage from "$lib/components/StatusMessage.svelte";
    import { GotoReload } from "$lib/functions/navigation";
    import type { StatusMessageData } from "$lib/types/types";

    let formStatus = $state<StatusMessageData>({
        loading: false,
        success: false,
        message: ""
    })

    async function handleSubmit(event: SubmitEvent) {
        formStatus = {
            loading: true,
            success: false,
            message: ""
        }
        const form = event.target as HTMLFormElement
        const formData = new FormData(form)
        try {
            const response = await fetch(`${API}/auth/register`, {
                method: "POST",
                body: formData,
                credentials: "include",
            })
            if (!response.ok) {
                var message = "an unknown error occurred"
                const data = await response.json()
                if (data.errcode) {
                    switch (data.errcode) {
                        case "ACCOUNT_USERNAME_EXISTS":
                            message = "an account with that username already exists"
                            break
                        case "ACCOUNT_EMAIL_EXISTS":
                            message = "an account with that email address already exists"
                            break
                    }
                }
                formStatus = {
                    loading: false,
                    success: false,
                    message: message
                }
                return
            } else {
                formStatus = {
                    loading: false,
                    success: true,
                    message: "registered!"
                }
                setTimeout(() => { GotoReload("/") }, 1000)
            }
            
        } catch (e) {
            formStatus = {
                loading: false,
                success: false,
                message: "error registering"
            }
        }
    }
</script>

<div id="title">
    <h2>register</h2>
    {#if formStatus.loading}
        <Loader />
    {/if}
</div>

<form onsubmit={handleSubmit}>
    <label>email <br />
        <input type="email" name="email" required>
    </label> <br />
    <label>username <br />
        <input type="username" name="username" required>
    </label> <br />
    <label>password <br />
        <input type="password" name="password" required>
    </label> <br />
    <button type="submit" disabled={formStatus.loading}>register</button>
</form>

<br />

<StatusMessage data={formStatus} />

<style>
    #title {
        display: flex;
        flex-direction: row;
        align-items: center;
    }

    #title>h2 {
        margin-right: 1rem;
    }

    form {
        display: block;
        flex-direction: column;
        text-align: left;
        width: fit-content;
    }

    button {
        width: fit-content;
        margin-left: auto;
        margin-top: 0.5rem;
    }
</style>