<script lang="ts">
    import type { Card } from '$lib/types/types.js';
    import { onMount } from 'svelte';

    let {data} = $props()

    interface Question {
        question: string
        answer: string
        response: string
        correct: boolean
    }

    let questions = $state<Question[]>([])
    
    let mode = $state("setup")

    function shuffle(array: any[]) {
        for (let i = array.length - 1; i > 0; i--) {
            const j = Math.floor(Math.random() * (i + 1));
            [array[i], array[j]] = [array[j], array[i]];
        }
        return array;
    }

    function createTest() {
        if (data.set?.cards) {
            questions = []
            const cards = data.set.cards
            for (let i = 0; i < cards.length; i++) {
                if (typeof cards[i].front == 'string' && typeof cards[i].back == 'string') {
                    questions.push({
                        question: String(cards[i].front),
                        answer: String(cards[i].back),
                        response: "",
                        correct: false
                    })
                }
            }
            questions = shuffle(questions)
            mode = "test"
        }
    }

    function gradeTest() {
        questions.forEach((q) => {
            q.correct = (q.response.trim().toLowerCase() == q.answer.trim().toLowerCase())
        })
        mode = "review"
    }
</script>

<a href={`/sets/${data.set?.id}`}>back to set</a>
<h2>test</h2>
{#if mode == "setup"}
    <form>
        <button onclick={createTest}>create test</button>
    </form>
{:else if mode == "test"}
    {#each questions as q}
        <p class="question">{q.question}</p>
        <div class="response" contenteditable="true" bind:innerText={q.response} role="input"></div>
    {/each}
    <br />
    <button onclick={gradeTest}>submit</button>
{:else if mode == "review"}
    {#each questions as q}
        <div class={`question_review ${q.correct ? "correct" : "incorrect"}`}>
            <p class="question">question: {q.question}</p>
            <p class={`response ${q.correct ? "correct" : "incorrect"}`}>your answer: {q.response}</p>
            <p class="answer">correct answer: {q.answer}</p>
        </div>
    {/each}
    <button onclick={createTest}>retake test</button>
{:else}
    <p>idk something broke</p>
{/if}

<style>
    .question {
        word-wrap: break-word;
    }

    .response.correct, .answer.correct {
        color: var(--col-msg-success);
    }

    .response.incorrect {
        color: var(--col-msg-error);
    }

    .answer.incorrect {
        color: var(--col-msg-neutral);
    }

    .question_review {
        padding: 1rem;
        margin-bottom: 0.5rem;
    }

    .question_review.correct {
        border: 2px solid var(--col-msg-success);
    }

    .question_review.incorrect {
        border: 2px solid var(--col-msg-error);
    }

</style>

