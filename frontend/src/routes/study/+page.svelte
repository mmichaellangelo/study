<script lang="ts">

    let { data } = $props()

</script>

<div id="title">
    <h2>my study sets</h2>
    <button>create a new set</button>
</div>

<div id="set_list">
    {#if data.sets}
        {#each data.sets as set}
            <a href={`/sets/${set.id}`} class="set">
                <span class="set_name">{set.name || "(untitled)"}</span>
                <span class="set_date">created {new Date(set.created).toLocaleDateString()}</span>
            </a>
        {/each}
    {/if}
</div>

<style>
    #title {
        display: flex;
        flex-direction: row;
        align-items: center;
        justify-content: space-between;
        max-width: 40rem;
    }

    #set_list {
        display: flex;
        flex-direction: column;
        max-width: 40rem;
        padding: 1rem;
    }

    .set {
        display: flex;
        justify-content: space-between;
        flex-direction: row;
        align-items: left;
        padding: 0.7rem;
        margin-bottom: 0.5rem;
        border-radius: 0.5rem;
        background-color: var(--col-purplegrey);
        color: var(--col-lightpink);
    }

    .set:hover {
        color: var(--col-lightblue);
        outline: 2px solid var(--col-lightpink);
    }

    button {
        width: fit-content;
        margin-left: auto;
        margin-right: 1rem;
        margin-bottom: 0.5rem;
    }
</style>
