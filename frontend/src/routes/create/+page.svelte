<script lang="ts">
    import { goto } from "$app/navigation";
    import { API } from "$lib/api";
    import Loader from "$lib/components/Loader.svelte";
    import StatusMessage from "$lib/components/StatusMessage.svelte";
    import type { StatusMessageData } from "$lib/types/types";
    import { onMount } from "svelte";

    let status = $state<StatusMessageData>({
        loading: true,
        message: "",
        success: false
    })
    onMount(async () => {
        const res = await fetch(`${API}/sets/`, {
            method: "POST",
            credentials: "include",
        })
        if (!res.ok) {
            status = {
                loading: false,
                message: await res.text(),
                success: false
            }
        }
        const data = await res.json()
        status.loading = false
        goto(`/sets/${data.id}/edit`)
    })
</script>
{#if status.loading}
    <Loader />
{:else}
    <StatusMessage data={status}/>
{/if}