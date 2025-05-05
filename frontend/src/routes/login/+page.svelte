<h2>Login</h2>

<form action="http://localhost:8080/login" method="post">
    <label>Email or Username
        <input type="text" name="emailorusername">
    </label> <br />
    <label>Password
        <input type="password" name="password">
    </label> <br />
    <button type="submit">Log In</button>
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