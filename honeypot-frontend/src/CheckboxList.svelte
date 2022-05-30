<script lang="ts" context="module">
  export type Option = {
    label?: string;
    value: string;
  };
  let uniqueId = Math.random().toString(36).substring(2);
</script>

<script lang="ts">
  export let options: Option[];
  export let value: string[] = [];
  export let onInput: (value: string[]) => void;
</script>

<div class="checkbox-list">
  {#each options as option}
    <div class="option">
      <input
        type="checkbox"
        id={uniqueId + option.value}
        checked={value.includes(option.value)}
        on:change={() => {
          if (value.includes(option.value)) {
            value = value.filter((v) => v !== option.value);
          } else {
            value = [...value, option.value];
          }
          onInput(value);
        }}
      />
      <label for={uniqueId + option.value}>{option.label || option.value}</label
      >
    </div>
  {/each}
</div>

<style lang="scss">
  .checkbox-list {
    margin: 1rem;
    display: flex;
    flex-wrap: wrap;

    align-items: center;
    .option {
      display: flex;
      flex-direction: row;
      margin-right: 1rem;
      align-items: baseline;
      label {
        align-self: baseline;
        margin-bottom: 8px;
      }
    }
  }
</style>
