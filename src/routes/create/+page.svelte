<script lang="ts">
    import type { Card, Set } from "$lib/types/types";

    const blankCard: Card = {front: "", back: ""}
    let set = $state<Set>({name: "", cards: [blankCard]});

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

<h2>create a set</h2>

<div id="create_frame">
    <form>
        <input type="text" placeholder="title" bind:value={set.name}>
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
    
</style>