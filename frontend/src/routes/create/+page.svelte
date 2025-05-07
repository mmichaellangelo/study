<script lang="ts">
    import Loader from "$lib/components/Loader.svelte";
    import type { Card, Set } from "$lib/types/types";
    import { onMount } from "svelte";

    let isLoading = $state(false)

    const blankCard: Card = {front: "", back: ""}

    let set = $state<Set>({name: "", cards: [blankCard]});

    let created = $state(false)
    
    let titleTimeout = 0

    onMount(async () => {
        const res = CreateSet()
    })

    async function handleUpdateTitle(e: Event) {
        const inputElement = e.target as HTMLInputElement
        const newTitle = inputElement.value
        titleTimeout = Date.now()
        setTimeout(() => {
            if (Date.now() - titleTimeout > 800) {
                
            }
        }, 1000)
    }

    async function UpdateTitle() {
        if (!created) {
            
        }
    }

    async function CreateSet() {
        try {
            const res = await fetch("http://localhost:8080/sets", {
                method: "POST",
                credentials: "include",
            })
        } catch (e) {
            return Promise.reject()
        }
    }
    
    function addCard() {
        set.cards.push({...blankCard})
        // Update database
    }

    function removeCard(index: number) {
        set.cards = set.cards.filter((_, i) => i !== index)
        // Update database
    }

    let draggedIndex: number

    function handleDragStart(event: DragEvent, index: number) {
        draggedIndex = index;
        event.dataTransfer?.setData("text/plain", String(index));

    }

    function handleDragOver(event: DragEvent) {
    event.preventDefault(); // Necessary for allowing a drop
  }

  function handleDrop(event: DragEvent, dropIndex: number) {
    event.preventDefault(); // Prevent default browser drop behavior

    if (draggedIndex === dropIndex) return; // Don't do anything if dropped on itself

    // 1. Remove the dragged item from its original position
    const draggedItem = set.cards.splice(draggedIndex, 1)[0];

    // 2. Insert the dragged item into its new position
    set.cards.splice(dropIndex, 0, draggedItem);

    set = { ...set }; // Force Svelte to recognize the change and update the UI.  Important!
  }


</script>
<div id="title">
    <h2>create a set</h2>
    {#if isLoading}
        <Loader />
    {/if}
</div>

<div id="create_frame">
    <form>
        <input type="text" placeholder="title" bind:value={set.name} onchange={handleUpdateTitle}>
        <br />
            {#each set.cards as card, index}
            <div class="card" draggable="true"
                role="listitem"
                ondragstart={(event) => handleDragStart(event, index)}
                ondragover={handleDragOver}
                ondrop={(event) => handleDrop(event, index)}>
                    <span>{`${index + 1}. `}</span>
                    <input type="text" bind:value={card.front} placeholder="front">
                    <input type="text" bind:value={card.back} placeholder="back">
                    <button onclick={() => removeCard(index)}>-</button>
            </div>
            {/each}
        
        <button onclick={addCard}>+</button>
    </form>
</div>

<style>
    #title {
        display: flex;
        flex-direction: row;
        align-items: center;
    }

    #title>h2 {
        margin-right: 1rem;
    }
</style>