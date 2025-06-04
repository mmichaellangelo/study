<svelte:head>
    <title>disco - flashcards: {data.set?.name}</title>
</svelte:head>

<script lang="ts">
    import type { Card } from '$lib/types/types.js';
    import { onMount } from 'svelte';

    let {data} = $props()
    let flipped = $state(false)
    let direction = $state("")
    let cardIndex = $state(0)
    let hideBack = $state(false)
    let cardsLength = $derived(data.set?.cards?.length)
    let currentCard: Card | undefined = $derived.by(() => {
        if (data.set?.cards) {
            return data.set.cards[cardIndex]
        } else {
            return undefined
        }
    })
    
    onMount(() => {
        document.removeEventListener("keydown", handleKey)
        if (data.set?.cards) {
            currentCard = data.set?.cards[0]
        }
        document.addEventListener("keydown", handleKey)
    })

    function handleKey(e: KeyboardEvent) {
        console.log(e.code)
        switch (e.code) {
            case 'Space':
                flip()
                break
            case 'ArrowUp':
                flip()
                break
            case 'ArrowDown': 
                flip()
                break
            case 'ArrowLeft':
                prev()
                break
            case 'ArrowRight':
                next()
                break
        }
            
    }

    function flip() {
        flipped = !flipped
    }

    function next() {
        if (cardsLength) {
            flipped = false
            direction = "from_right"
            hideBack = true
            setTimeout(() => {
                direction = ""
                hideBack = false
            }, 100)
            if (cardIndex < (cardsLength - 1)) {
                cardIndex++
            } else {
                cardIndex = 0
            }
        }
    }

    function prev() {
        if (cardsLength) {
            flipped = false
            direction = "from_left"
            hideBack = true
            setTimeout(() => {
                direction = ""
                hideBack = false
            }, 100)
            if (cardIndex > 0) {
                cardIndex--
            } else {
                cardIndex = cardsLength - 1
            }
        }
    }

    var touchStartX = 0;
    var touchStartY = 0;
    var touchEndX = 0;
    var touchEndY = 0;

    function touchRestart() {
        touchStartX = 0;
        touchStartY = 0;
        touchEndX = 0;
        touchEndY = 0;
    }

    function handleTouchStart(e: TouchEvent) {
        touchStartX = e.changedTouches[0].screenX;
        touchStartY = e.changedTouches[0].screenY;
    }

    function handleTouchEnd(e: TouchEvent) {
        touchEndX = e.changedTouches[0].screenX;
        touchEndY = e.changedTouches[0].screenY;

        if (touchEndX - touchStartX > 50) {
            prev();
            touchRestart();
        }
        if (touchEndX - touchStartX < -50) {
            next();
            touchRestart();
        }
    }
</script>

<div id="container">
{#if data.set && currentCard}

    <a href={`/sets/${data.set.id}`}>back to set</a>
    <h2>{data.set.name}</h2>

    <div id="navigation">
        <button onclick={prev}>prev</button>
        <span>{cardIndex + 1} of {cardsLength}</span>
        <button onclick={next}>next</button>        
    </div>

    <div id="cards_container" onclick={flip} ontouchstart={handleTouchStart} ontouchend={handleTouchEnd} role="input" onkeydown={handleKey}>
        <div class={`card ${flipped ? "back" : "front"} ${direction}`}
             style={`border-color: ${flipped ? "var(--col-greenblue)" : "var(--col-lightpink"}`}>
            <h3>{currentCard.front}</h3>
        </div>
        <div class={`card ${flipped ? "front" : "back"} ${hideBack ? "hide" : ""}`} 
             style={`border-color: ${flipped ? "var(--col-greenblue)" : "var(--col-lightpink"};
                     background-color: ${flipped ? "var(--col-darkblue)" : "var(--col-purplegrey)"};`}>
            <p>{currentCard.back}</p>
        </div>
    </div>

{:else}
    <p>there was an error loading the set</p>
{/if}
</div>

<style>
    #container {
        position: relative;
    }

    #cards_container {
        margin-top: 1rem;
        perspective: 1000px;
        background-color: transparent;
        border: none;
        display: flex;
        flex-direction: column;
        align-items: center;
        width: 100%;
    }

    #navigation {
        width: 100%;
        display: flex;
        flex-direction: row;
        justify-content: center;
        align-items: center;
    }

    #navigation>span {
        margin-left: 0.5rem;
        margin-right: 0.5rem;
    }

    .card {
        word-wrap: break-word;
        padding: 1rem;
        backface-visibility: hidden;
        position: absolute;
        cursor: pointer;
        display: flex;
        flex-direction: row;
        align-items: center;
        text-align: center;
        width: 90%;
        aspect-ratio: 5 / 3;
        max-width: 30rem;
        background-color: var(--col-purplegrey);
        border-radius: 1rem;
    }

    .card>h3, .card>p {
        width: 100%;
    }

    .card>p {
        color: var(--col-greenblue);
    }

    .front {
        transform: rotateX(0deg);
        transition: 150ms;
        border: 2px solid;
    }

    .back {
        transform: rotateX(180deg);
        transition: 150ms;
        border: 2px solid;
    }

    .hide {
        transition: 0ms;
        opacity: 0;
    }

    @keyframes from_left {
        from {
            transform: rotateY(20deg) translateX(-2rem);
        }
    }

    @keyframes from_right {
        from {
            transform: rotateY(-20deg) translateX(2rem);
        }
    }

    .from_left {
        transition: 0ms;
        animation: 100ms forwards from_left;
    }

    .from_right {
        transition: 0ms;
        animation: 100ms forwards from_right;
    }

</style>