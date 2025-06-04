<svelte:head>
    <title>disco - test: {data.set?.name}</title>
</svelte:head>
<script lang="ts">

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

    let numQuestions = $state(data.set?.cards?.length || 1)
    let answerWith = $state("front")
    let order = $state("random")

    function createTest() {
        if (data.set?.cards) {
            questions = []
            const cards = data.set.cards
            for (let i = 0; i < numQuestions; i++) {
                if (typeof cards[i].front == 'string' && typeof cards[i].back == 'string') {
                    if (answerWith == "front") {
                        questions.push({
                            question: String(cards[i].back),
                            answer: String(cards[i].front),
                            response: "",
                            correct: false
                        })
                    } else {
                        questions.push({
                            question: String(cards[i].front),
                            answer: String(cards[i].back),
                            response: "",
                            correct: false
                        })
                    }
                }
            }
            switch (order) {
                case "random":
                    questions = shuffle(questions)
                    break
                case "created":
                    break
                case "reverse":
                    questions.reverse()
                    break
            }   
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
    <label>
        number of questions (set has {data.set?.cards?.length} cards) <br />
        <input type="number" min="1" max={data.set?.cards?.length} defaultvalue={data.set?.cards?.length} bind:value={numQuestions} /> <br />
    </label> <br />
    <label>
        answer with: <br />
        <label>
            <input type="radio" name="answer_with" value="front" bind:group={answerWith}>
            front
        </label> <br />
        <label>
            <input type="radio" name="answer_with" value="back" bind:group={answerWith}>
            back
        </label> <br />
    </label> <br />
    <label>
        order: <br />
        <label>
            <input type="radio" name="order" value="random" bind:group={order}>
            random
        </label> <br />
        <label>
            <input type="radio" name="order" value="created" bind:group={order}>
            created
        </label> <br />
        <label>
            <input type="radio" name="order" value="reverse" bind:group={order}>
            reverse
        </label> <br />
    </label> <br />
    <button onclick={createTest}>create test</button>
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

