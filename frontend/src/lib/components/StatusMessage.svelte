<script lang="ts">
    import type { StatusMessageData } from "$lib/types/types";

    let { data } = $props<{data: StatusMessageData}>()

    let color = $derived.by(() => {
        if (data.loading) {
            return "var(--col-msg-neutral)"
        } else if (data.success) {
            return "var(--col-msg-success)"
        } else {
            return "var(--col-msg-error)"
        }
    })
</script>

<span style={`color: ${color};`}>{data.message}</span>

