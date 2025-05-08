<script lang="ts">

    import Header from "$lib/components/Header.svelte";
    import { userState } from "$lib/state/account.svelte.js";

    import "$lib/styles/global.css"
    import { onMount } from "svelte";

	let { children } = $props()


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
