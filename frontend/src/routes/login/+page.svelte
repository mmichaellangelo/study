<script lang="ts">
    import { GotoReload } from "$lib/functions/navigation";

    let isLoading = $state(false)
    let errorMessage = $state("")

    async function handleSubmit(event: SubmitEvent) {
        isLoading = true
        const form = event.target as HTMLFormElement
        const formData = new FormData(form)

        try {
            const response = await fetch("http://localhost:8080/login", {
                method: "POST",
                body: formData,
                credentials: "include"
            })
            if (!response.ok) {
                errorMessage = await response.text()
                return
            }
            GotoReload("/")
        } catch (e: any) {
            errorMessage = e.toString()
        }
    }

</script>
<h2>login</h2>

<form onsubmit={handleSubmit}>
    <label>Email or Username
        <input type="text" name="emailorusername">
    </label> <br />
    <label>Password
        <input type="password" name="password">
    </label> <br />
    <button type="submit" disabled={isLoading}>Log In</button>
</form>
{#if isLoading}
<p>Logging in...</p>
{/if}

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