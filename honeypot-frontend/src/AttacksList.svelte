<script lang="ts">
  import { fetchAPI, WebSocketAPI } from "./api";
  import Badge from "./Badge.svelte";
  import CheckboxList from "./CheckboxList.svelte";

  import Error from "./Error.svelte";

  import Loader from "./Loader.svelte";

  interface IAttack {
    ID: number;
    CreatedAt: string;
    UpdatedAt: string;
    DeletedAt: string | null;
    protocol: string;
    inProgress: boolean;
    sourceIP: string;
    country: string;
    countryCode: string;
    location: string;
    isp: string;
    contents: string;
    duration: number;
    classification: string;
  }

  let propertyLabels = {
    ID: "ID",
    CreatedAt: "Started",
    inProgress: "In progress",
    protocol: "Protocol",
    sourceIP: "Source IP",
    country: "Country",
    countryCode: "Country code",
    location: "Location",
    duration: "Duration",
    classification: "classification",
  };

  let propertyTypes = {
    CreatedAt: "date",
    duration: "ms",
  };

  let classificationOptions = [
    { value: "empty" },
    { value: "username_entered" },
    { value: "authenticated" },
    { value: "command_entered" },
  ];
  let protocolsOptions = [{ value: "telnet" }, { value: "ssh" }];
  let protocols = ["telnet", "ssh"];
  let classifications = [
    "username_entered",
    "authenticated",
    "command_entered",
  ];
  function formatDate(d: string) {
    return new Date(d).toLocaleString();
  }

  function formatDuration(d: number) {
    let str = "";
    let msec = d % 1000;
    let sec = Math.floor(d / 1000);
    let min = Math.floor(sec / 60);
    let hour = Math.floor(min / 60);

    if (hour > 0) {
      str += `${hour}h `;
    }
    if (min > 0) {
      str += `${min}m `;
    }
    if (sec > 0) {
      str += `${sec}s `;
    }
    if (msec > 0) {
      str += `${msec}ms `;
    }
    return str;
  }

  let attacks: IAttack[] = [];
  let loadingPromise: Promise<any> | null = null;
  function loadMoreAttacks() {
    loadingPromise = fetchAPI("GET", "/api/attacks", undefined, {
      classifications,
      protocols,
    }).then((data) => {
      attacks = attacks.concat(data);
    });
  }
  loadMoreAttacks();

  function classificationsChanged(c: string[]) {
    classifications = c;
    attacks = [];
    loadMoreAttacks();
  }

  function protocolsChanged(p: string[]) {
    protocols = p;
    attacks = [];
    loadMoreAttacks();
  }

  let attacksWebSocket: WebSocket = WebSocketAPI("/api/attacks/ws");
  attacksWebSocket.onmessage = (e) => {
    let data = JSON.parse(e.data);
    // replace attack with the same id or unshift new attack
    let attack = attacks.find((a) => a.ID === data.ID);
    if (attack) {
      attacks = attacks.map((a) => (a.ID === data.ID ? data : a));
    } else {
      attacks = [data, ...attacks];
    }
    // sort by ids in descending order
    attacks.sort((a, b) => b.ID - a.ID);
    attacks = attacks.filter((a) => {
      if (a.inProgress) {
        return true;
      }
      return classifications.includes(a.classification) && protocols.includes(a.protocol);
    });
  };
</script>

<div>
  Classification:
  <CheckboxList
    options={classificationOptions}
    value={classifications}
    onInput={classificationsChanged}
  />
  Protocols:
  <CheckboxList
    options={protocolsOptions}
    value={protocols}
    onInput={protocolsChanged}
  />
  {#each attacks as attack}
    <div class="attack">
      {#if attack.inProgress}
        <Badge>In progress</Badge>
      {/if}
      <div class="columns">
        <div class="properties">
          {#each Object.entries(propertyLabels) as [key, label]}
            <div class="property">
              <label>{label}</label>
              <div class="value">
                {#if attack[key] === null}
                  <span class="null">null</span>
                {:else if propertyTypes[key] === "date"}
                  {formatDate(attack[key])}
                {:else if propertyTypes[key] === "ms"}
                  {formatDuration(attack[key])}
                {:else}
                  {attack[key]}
                {/if}
              </div>
            </div>
          {/each}
        </div>
        <div class="contents">
          <pre>{attack.contents}</pre>
        </div>
      </div>
    </div>
  {/each}
  {#await loadingPromise}
    <Loader />
  {:catch error}
    <Error {error} />
  {/await}
</div>

<style lang="scss">
  .attack {
    background: #2d3436;
    padding: 1rem;
    margin-bottom: 1rem;
    .columns {
      display: flex;
      flex-direction: row;
      @media (max-width: 900px) {
        flex-direction: column;
      }
      .properties {
        padding-top: 16px;
        flex: 1;
        display: flex;
        flex-direction: column;
        .property {
          margin-bottom: 0.8rem;
          label {
            color: var(--primary);
            font-size: 0.8rem;
          }
        }
      }
      .contents {
        flex: 1;
        background: #636e72;
        pre {
          margin: 1rem;
          max-height: 500px;
          overflow-y: auto;
        }
      }
    }
  }
</style>
