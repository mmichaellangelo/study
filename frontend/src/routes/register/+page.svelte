<script lang="ts">
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
            const response = await fetch("http://localhost:8080/register", {
                method: "POST",
                body: formData,
                credentials: "include",
            })
            if (!response.ok) {
                formStatus = {
                    loading: false,
                    success: false,
                    message: await response.text()
                }
                return
            } else {
                formStatus = {
                    loading: false,
                    success: true,
                    message: "registration success! redirecting..."
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
    <button type="submit" disabled={formStatus.loading}>Register</button>
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