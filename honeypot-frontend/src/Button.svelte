<button {...$$props} on:click>
  <slot />
</button>

<style lang="scss">
  button {
    width: 100%;
    padding: 8px;
    padding-bottom: 4px;;
    color: var(--primary);
    background: transparent;
    border: 1px solid var(--primary);
    outline: none;
    cursor: pointer;
    font-family: monospace;
    vertical-align: center;
    text-align: center;
    font-weight: bold;
    text-transform: uppercase;
    &:hover {
        background: var(--primary);
        color: white;
    }
    &:active {
        transform: scale(0.9)
     
    }
  }
</style>
