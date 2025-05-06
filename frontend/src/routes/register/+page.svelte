<script lang="ts">
    async function handleSubmit(event: SubmitEvent) {
        const form = event.target as HTMLFormElement
        const formData = new FormData(form)
        // try {
        //     const response = await fetch("http://localhost:8080/register", {
        //         method: "POST",
        //         body: formData,
        //         credentials: "include",
        //     })
        //     if (!response.ok {
                
        //     })
        // }
    }
</script>

<h2>register</h2>

<form onsubmit={handleSubmit}>
    <label>Email
        <input type="email" name="email">
    </label> <br />
    <label>Username
        <input type="username" name="username">
    </label> <br />
    <label>Password
        <input type="password" name="password">
    </label> <br />
    <button type="submit">Register</button>
</form>

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