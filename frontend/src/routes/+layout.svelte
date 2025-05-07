<script lang="ts">

    import Header from "$lib/components/Header.svelte";
    import { userState } from "$lib/state/account.svelte.js";

    import "$lib/styles/global.css"
    import { onMount } from "svelte";

	let { children, data } = $props()

    onMount(async () => {
        if (userState.ID == -1) {
        try {
            const res = await fetch(`http://localhost:8080/me`,
                {
                    method: "GET",
                    credentials: "include",
                }
            )
            const data = await res.json()
            
            userState.ID = data.userid
            userState.Username = data.username
            
        } catch (e) {
            userState.ID = -1
            userState.Username = ""
        }
    }
})


</script>
<Header />

<div id="page_body">
    {@render children?.()}
</div>

<style>
    #page_body {
        padding: 1rem;
    }
</style>
