<!-- CODE FROM https://cssloaders.github.io/ (thank you!) -->

<span class="loader"></span>

<style>
    .loader {
    width: 1rem;
    height: 1rem;
    border: 3px solid var(--col-lightblue);
    border-bottom-color: transparent;
    border-radius: 50%;
    display: inline-block;
    box-sizing: border-box;
    animation: rotation 1s linear infinite;
    }

    @keyframes rotation {
    0% {
        transform: rotate(0deg);
    }
    100% {
        transform: rotate(360deg);
    }
    } 
</style>