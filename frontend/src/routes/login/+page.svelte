<script lang="ts">
    import { API } from "$lib/api";
    import Loader from "$lib/components/Loader.svelte";
    import StatusMessage from "$lib/components/StatusMessage.svelte";
    import { GotoReload } from "$lib/functions/navigation";
    import type { StatusMessageData } from "$lib/types/types";

    let formState = $state<StatusMessageData>({
        loading: false,
        message: "",
        success: false,
    })

    async function handleSubmit(event: SubmitEvent) {
        formState = {
            loading: true,
            message: "",
            success: false
        }
        const form = event.target as HTMLFormElement
        const formData = new FormData(form)

        try {
            const response = await fetch(`http://${API}/login`, {
                method: "POST",
                body: formData,
                credentials: "include"
            })
            if (!response.ok) {
                formState = {
                    loading: false,
                    message: await response.text(),
                    success: false,
                }
            } else {
                formState = {
                    loading: false,
                    message: `login success! redirecting...`,
                    success: true,
                }
                setTimeout(() => { GotoReload("/") }, 1000)
            }
            
        } catch (e: any) {
            formState = {
                loading: false,
                message: "an unknown error occurred",
                success: false,
            }
            return
        }
    }

</script>

<div id="title">
    <h2>login</h2>
    {#if formState.loading}
        <Loader />
    {/if}
</div>

<form onsubmit={handleSubmit}>
    <label>email or username <br />
        <input type="text" name="emailorusername" required>
    </label> <br />
    <label>password <br />
        <input type="password" name="password" required>
    </label> <br />
    <button type="submit" disabled={formState.loading}>Log In</button>
</form>

<br />

<StatusMessage data={formState} />

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
        display: flex;
        flex-direction: column;
        text-align: right;
        width: fit-content;
    }
    button {
        width: fit-content;
        margin-left: auto;
    }
</style>