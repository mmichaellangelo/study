<script lang="ts">
    let { data } = $props()
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
    <p>there was an error loading the set{data.error? `: ${data.error}` : ""}</p>
{/if}

<style>
    th, td {
        padding: 0.5rem;
        outline: 1px solid var(--col-lightblue);
    }
</style>