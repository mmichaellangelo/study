<script lang="ts">
    let { data } = $props()
</script>

<h2>study</h2>

{#if data.sets}
    {#each data.sets as set}
        <p>{set.id}</p>
    {/each}
{/if}
