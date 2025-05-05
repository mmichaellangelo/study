<script lang="ts">

    import Header from "$lib/components/Header.svelte";
    import { userState } from "$lib/state/account.svelte.js";

    import "$lib/styles/global.css"
    import { onMount } from "svelte";

	let { children, data } = $props()

    onMount(() => {
        userState.ID = data.userID
        console.log("setting userID to " + data.userID)
    })


</script>
<Header />

{@render children?.()}