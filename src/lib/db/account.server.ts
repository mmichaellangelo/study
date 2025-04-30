import { hash } from "bcrypt";
import { supabase } from "./supabaseClient.server";

function emailIsValid(email: string) {
    return /^[^@\s]+@[^@\s]+\.[^@\s]+$/.test(email);
}
  
async function createUser(username: string, email: string, password: string) {
    if (username.trim().length == 0) {
        throw new Error("empty username")
    }
    if (!emailIsValid(email)) {
        throw new Error("invalid email")
    }

    const session = await supabase.auth.signUp({
        email: email,
        password: password,
    })

}