<script lang="ts">
    import { goto } from '$app/navigation';
    import Loader from '$lib/components/Loader.svelte';
    import { onMount } from 'svelte';

    let { data } = $props()
    let isLoading = $state(true)
    let showBack = $state(false)
    let errorMessage: string = $derived.by(() => {
        if (!data.error) {
            return ""
        }
        switch (data.error.errcode) {
            case "ACCESS_NOT_ALLOWED":
                return "sorry, you don't have access to this set."
            default:
                return "an unknown error occurred."
        }
    })

    onMount(() => {
        isLoading = false
    })
</script>

<a href="/study">back to study sets</a>

{#if data.set}
    <h2>{data.set.name}</h2>
    <button onclick={() => goto(`/sets/${data.set.id}/edit`)}>edit</button>
    <button onclick={() => goto(`/sets/${data.set.id}/flashcards`)}>flashcards</button>
    <button onclick={() => goto(`/sets/${data.set.id}/test`)}>test</button>

    <br /> <br />
    {#if data.set.cards}
    <table>
        <thead>
            <tr>
                <th>front</th>
                <th>
                    {#if showBack}
                        back
                    {:else}
                        <button onclick={() => showBack = true}>show back</button>
                    {/if}
                    
                </th>
            </tr>
        </thead>
        <tbody>
            {#each data.set.cards as card}
                <tr>
                    <td>{card.front}</td>
                    <td>
                        {#if showBack}
                            {card.back}
                        {:else}
                            ---
                        {/if}
                    </td>
                </tr>
            {/each}
        </tbody>
    </table>
    {/if}
{:else}
    {#if isLoading}
        <Loader />
    {:else}
        <p>{errorMessage}</p>
    {/if}
{/if}

<style>

    table {
        width: 100%;
        max-width: 30rem;
        border-collapse: collapse;
        table-layout: fixed;
    }

    th, td {
        padding: 0.5rem;
        outline: 1px solid var(--col-lightblue);
        word-wrap: break-word;
    }

    th {
        text-align: left;
        color: var(--col-lightblue);
    }
</style>