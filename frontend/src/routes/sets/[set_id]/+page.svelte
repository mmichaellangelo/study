<script lang="ts">
    let { data } = $props()
</script>

{#if data.set}
    <h3>{data.set.name}</h3>
    {#if data.set.cards}
        {#each data.set.cards as card}
            <div class="card">
                <span>{card.front}</span>
                <span>{card.back}</span>
            </div>
        {/each}
    {/if}
{/if}