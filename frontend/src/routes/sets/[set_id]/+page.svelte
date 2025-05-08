<script lang="ts">
    import Loader from '$lib/components/Loader.svelte';
    import { onMount } from 'svelte';

    let { data } = $props()
    let isLoading = $state(true)

    onMount(() => {
        isLoading = false
    })
</script>

<button onclick={() => history.back()}>back</button>

{#if data.set}
    <h3>{data.set.name}</h3>
    <a href={`/sets/${data.set.id}/edit`}>edit set</a>
    <br /> <br />
    {#if data.set.cards}
    <table>
        <thead>
            <tr>
                <th>id</th>
                <th>front</th>
                <th>back</th>
            </tr>
        </thead>
        <tbody>
            {#each data.set.cards as card}
                <tr>
                    <td>{card.id}</td>
                    <td>{card.front}</td>
                    <td>{card.back}</td>
                </tr>
            {/each}
        </tbody>
    </table>
    {/if}
{:else}
    {#if isLoading}
        <Loader />
    {:else}
        <p>there was an error loading the set{data.error? `: ${data.error}` : ""}</p>
    {/if}
{/if}

<style>
    th, td {
        padding: 0.5rem;
        outline: 1px solid var(--col-lightblue);
    }
</style>