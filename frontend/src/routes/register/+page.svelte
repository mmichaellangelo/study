<script lang="ts">
    import StatusMessage from "$lib/components/StatusMessage.svelte";
    import { GotoReload } from "$lib/functions/navigation";
    import type { StatusMessageData } from "$lib/types/types";

    let formStatus = $state<StatusMessageData>({
        loading: false,
        success: false,
        message: ""
    })

    async function handleSubmit(event: SubmitEvent) {
        console.log("submit")
        formStatus = {
            loading: true,
            success: false,
            message: "loading..."
        }
        const form = event.target as HTMLFormElement
        const formData = new FormData(form)
        try {
            const response = await fetch("http://localhost:8080/register", {
                method: "POST",
                body: formData,
                credentials: "include",
            })
            console.log(response)
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

<h2>register</h2>

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

<StatusMessage data={formStatus} />

<style>
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