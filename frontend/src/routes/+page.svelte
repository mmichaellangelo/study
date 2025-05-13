<script lang="ts">
    import { goto } from "$app/navigation";

</script>

<div id="parent">
    <div id="hero">
        <h2>studying should be fun!</h2>
        <span>disco is a free study app to help you learn anything</span>
        <p><a class="linkbutton" href="/register">make a free account</a> to join the party</p>
        <div id="hero_card_left"></div>
        <div id="hero_card_right"></div>
    </div>
</div>


<style>
    @keyframes herointro {
        from {
            border: 2px solid var(--col-purplegrey);
            box-shadow: 0px 0px 10px var(--col-lightpink);
            transform: rotateX(15deg) rotateY(10deg);
        }
    }

    #parent {
        perspective: 1000px;
    }

    #hero {
        perspective: 100px;
        position: relative;
        display: flex;
        flex-direction: column;
        align-items: center;
        background-color: transparent;
        width: fit-content;
        max-width: 70%;
        padding: 1rem;
        margin-left: auto;
        margin-right: auto;
        border-radius: 1rem;
        border: 2px solid var(--col-lightblue);
        animation: 0.5s ease-out forwards herointro;
        box-sizing: border-box;
        z-index: 0;
    }

    #hero::before {
        content: "";
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: var(--col-purplegrey);
        border-radius: inherit;
        z-index: -1;
    }

    @keyframes fadefromblue {
        from {
            color: var(--col-lightblue);
        }
    }

    #hero>h2 {
        color: var(--col-lightpink);
        animation: 0.5s fadefromblue;
    }

    @keyframes fadeup {
        from {
            opacity: 0;
            transform: translateY(1rem);
        }
    }

    #hero>span, #hero>p {
        animation: 0.5s fadeup;
    }

    #hero_card_left, #hero_card_right {
        padding: 0px;
        margin: 0px;
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background-color: transparent;
        border: 2px solid var(--col-lightpink);
        border-radius: 1rem;
        opacity: 0;
    }

    #hero_card_left {
        animation: 0.5s 0.1s ease forwards cardleft;
        z-index: -2;
    }

    #hero_card_right {
        animation: 0.5s 0.1s ease forwards cardright;
        z-index: -3;
    }

    @keyframes cardleft {
        to {
            translate: -1rem 1rem;
            rotate: -3deg;
            opacity: 0.5;
        }
    }

    @keyframes cardright {
        to {
            translate: 1rem 1rem;
            rotate: 2deg;
            opacity: 0.7;
        }
    }

</style>