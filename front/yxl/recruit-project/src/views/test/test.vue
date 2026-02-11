<template>
  <div class="charts-container">
    <div class="chart-wrapper">
      <h3>饼图示例</h3>
      <v-chart class="chart" :option="pieOption" />
    </div>
    <div class="chart-wrapper">
      <h3>柱状图示例</h3>
      <v-chart class="chart" :option="barOption" />
    </div>
  </div>
</template>

<script>
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { PieChart, BarChart } from "echarts/charts";
import {
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
} from "echarts/components";
import VChart, { THEME_KEY } from "vue-echarts";
import { ref, defineComponent } from "vue";

use([
  CanvasRenderer,
  PieChart,
  BarChart,
  TitleComponent,
  TooltipComponent,
  LegendComponent,
  GridComponent
]);

export default defineComponent({
  name: "HelloWorld",
  components: {
    VChart
  },
  provide: {
    [THEME_KEY]: "light"
  },
  setup: () => {
    const pieOption = ref({
      title: {
        text: "Traffic Sources",
        left: "center"
      },
      tooltip: {
        trigger: "item",
        formatter: "{a} <br/>{b} : {c} ({d}%)"
      },
      legend: {
        orient: "vertical",
        left: "left",
        data: ["Direct", "Email", "Ad Networks", "Video Ads", "Search Engines"]
      },
      series: [
        {
          name: "Traffic Sources",
          type: "pie",
          radius: "55%",
          center: ["50%", "60%"],
          data: [
            { value: 335, name: "Direct" },
            { value: 310, name: "Email" },
            { value: 234, name: "Ad Networks" },
            { value: 135, name: "Video Ads" },
            { value: 1548, name: "Search Engines" }
          ],
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: "rgba(0, 0, 0, 0.5)"
            }
          }
        }
      ]
    });

    const barOption = ref({
      title: {
        text: "月度销售数据",
        left: "center"
      },
      tooltip: {
        trigger: "axis",
        axisPointer: {
          type: "shadow"
        }
      },
      legend: {
        data: ["销量"],
        top: "10%"
      },
      grid: {
        left: "3%",
        right: "4%",
        bottom: "3%",
        containLabel: true
      },
      xAxis: {
        type: "category",
        data: ["衬衫", "羊毛衫", "雪纺衫", "裤子", "高跟鞋", "袜子"]
      },
      yAxis: {
        type: "value"
      },
      series: [
        {
          name: "销量",
          type: "bar",
          data: [120, 200, 150, 80, 70, 110],
          itemStyle: {
            color: "#5470c6"
          }
        }
      ]
    });

    return { pieOption, barOption };
  }
});
</script>

<style scoped>
.charts-container {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding: 20px;
}

.chart-wrapper {
  width: 100%;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  padding: 20px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart-wrapper h3 {
  margin: 0 0 15px 0;
  text-align: center;
  color: #333;
  font-size: 16px;
  font-weight: 600;
}

.chart {
  height: 400px;
}
</style>