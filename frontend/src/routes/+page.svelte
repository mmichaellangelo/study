<script lang="ts">
    import { goto } from "$app/navigation";

</script>

<div id="hero">
    <h2>studying should be fun!</h2>
    <span>disco is a free study app to help you learn anything</span>
    <p><a class="linkbutton" href="/register">make a free account</a> to join the party</p>
    <div id="hero_card_left"></div>
    <div id="hero_card_right"></div>
</div>


<style>
    @keyframes fun-border {
        0% {
            background-color: var(--col-darkblue);
            border-radius: 2rem;
            border: 2px solid var(--col-purplegrey);
            box-shadow: 0px 0px 10px var(--col-lightpink);
        }

        100% {
            border-radius: '';
            border: '';
        }
    }

    #hero {
        display: flex;
        flex-direction: column;
        align-items: center;
        background-color: var(--col-purplegrey);
        width: fit-content;
        padding: 1rem;
        margin-left: auto;
        margin-right: auto;
        border-radius: 1rem;
        border: 2px solid var(--col-lightblue);
        animation: 0.5s ease-out normal fun-border;
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

    @keyframes cardleft {
        to {
            transform: translateX(-5rem);
            rotate: -10deg;
            border-radius: 2rem;
            opacity: 0.5;
        }
    }

    #hero_card_left {
        position: absolute;
        z-index: -1;
        width: 20rem;
        height: 10rem;
        background-color: var(--col-lightblue);
        animation: 0.5s ease forwards cardleft;
    }

    @keyframes cardright {
        to {
            translate: 3rem 1rem;
            rotate: 7deg;
            border-radius: 2rem;
            opacity: 0.7;
        }
    }

    #hero_card_right {
        opacity: 0;
        position: absolute;
        z-index: -1;
        width: 20rem;
        height: 10rem;
        background-color: var(--col-lightblue);
        animation: 0.5s ease forwards cardright;
    }

</style>