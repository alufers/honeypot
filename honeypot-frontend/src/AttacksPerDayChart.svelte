<script lang="ts">
  import { fetchAPI } from "./api";
  import Error from "./Error.svelte";
  import Loader from "./Loader.svelte";
  import Chart from "svelte-frappe-charts";

  interface IAttackPerDay {
    time: string;
    count: number;
    protocol: string;
  }

  const chartDataPromise = fetchAPI<IAttackPerDay[]>(
    "GET",
    "/api/attacks/stats/by-time?timeFormat=%Y-%m-%d"
  ).then((attacksPerDay) => {
    let protocols = [];
    const map: {
      time: Date;
      countByProtocol: {
        [protocol: string]: number;
      };
    }[] = [];
    attacksPerDay.forEach((att) => {
      const mapEntry = {
        time: att.time,
        countByProtocol: {},
        totalCount: 0,
      };
      if (!map[att.time]) {
        map[att.time] = mapEntry;
      }
      map[att.time].countByProtocol[att.protocol] = att.count;
      map[att.time].totalCount += att.count;
      if (!protocols.includes(att.protocol)) {
        protocols.push(att.protocol);
      }
    });
    const N_DAYS = 90;
    const lastNDays = Array(N_DAYS)
      .fill(0)
      .map((_, i) => new Date(new Date().getTime() - i * 24 * 60 * 60 * 1000))
      .reverse();
    let chartData = {
      labels: lastNDays.map((d) => d.toLocaleDateString()),
      datasets: protocols.map((p) => ({
        name: p,
        values: lastNDays.map(
          (d) => map[d.toISOString().substring(0, 10)]?.countByProtocol[p] || 0
        ),
      })),
    };

    return chartData;
  });
</script>

<div>
  {#await chartDataPromise}
    <Loader />
  {:then data}
    <div class="chart-backdrop">
      <Chart
        {data}
        type="bar"
        title="Connections per day (last 90 days)"
        barOptions={{ stacked: true, spaceRatio: 0.5 }}
        colors={["#00b894", "#e84393" , "#0984e3"]}
        axisOptions={{
            xAxisMode: 'tick',
            xIsSeries: true,
        }}
      />
    </div>
  {:catch error}
    <Error {error} />
  {/await}
</div>

<style>
  .chart-backdrop {
    /* background: #2d3436; */
  }
</style>
