<h2>Register</h2>

<form action="http://localhost:8080/register" method="post">
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