<script lang="ts">
    let { data } = $props()
</script>

<h2>study</h2>

{#if data.sets}
    {#each data.sets as set}
        <a href={`/sets/${set.id}`} class="set">
            <span>{set.name || "(untitled)"}</span>
        </a>
    {/each}
{/if}

<style>
    .set {
        display: block;
        padding: 0.5rem;
        margin-bottom: 0.5rem;
        background-color: var(--col-purplegrey);
        color: var(--col-lightpink);
    }

    .set:hover {
        color: var(--col-lightblue);
        outline: 2px solid var(--col-lightpink);
    }
</style>
