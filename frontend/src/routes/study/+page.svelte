<svelte:head>
    <title>disco - study sets</title>
</svelte:head>

<script lang="ts">
    import { goto } from '$app/navigation';


    let { data } = $props()

</script>

<div id="title">
    <h2>my study sets</h2>
    <button onclick={() => goto("/create")}>create a new set</button>
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
    }

    .set {
        display: flex;
        justify-content: space-between;
        flex-direction: row;
        align-items: left;
        padding: 0.7rem;
        padding-left: 1.4rem;
        padding-right: 1.4rem;
        margin-bottom: 0.5rem;
        border-radius: 2rem;
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
        margin-bottom: 0.5rem;
    }
</style>
