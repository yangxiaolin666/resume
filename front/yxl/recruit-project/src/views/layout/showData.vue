<template>
  <div class="charts-container">
    <div class="chart-wrapper">
      <v-chart class="chart" :option="pieOption" />
    </div>
    <div class="chart-wrapper">
      <v-chart class="chart" :option="barOption" />
    </div>
    <div class="chart-wrapper">
      <v-chart class="chart" :option="areaOption" />
    </div>
  </div>
</template>

<script>
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { PieChart, BarChart, LineChart } from "echarts/charts";
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
  LineChart,
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

    const areaOption = ref({
      title: {
        text: "网站访问趋势",
        left: "center"
      },
      tooltip: {
        trigger: "axis",
        axisPointer: {
          type: "cross",
          label: {
            backgroundColor: "#6a7985"
          }
        }
      },
      legend: {
        data: ["邮件营销", "联盟广告", "视频广告", "直接访问", "搜索引擎"],
        top: "10%"
      },
      grid: {
        left: "3%",
        right: "4%",
        bottom: "3%",
        containLabel: true
      },
      xAxis: [
        {
          type: "category",
          boundaryGap: false,
          data: ["周一", "周二", "周三", "周四", "周五", "周六", "周日"]
        }
      ],
      yAxis: [
        {
          type: "value"
        }
      ],
      series: [
        {
          name: "邮件营销",
          type: "line",
          stack: "总量",
          areaStyle: {},
          emphasis: {
            focus: "series"
          },
          data: [120, 132, 101, 134, 90, 230, 210]
        },
        {
          name: "联盟广告",
          type: "line",
          stack: "总量",
          areaStyle: {},
          emphasis: {
            focus: "series"
          },
          data: [220, 182, 191, 234, 290, 330, 310]
        },
        {
          name: "视频广告",
          type: "line",
          stack: "总量",
          areaStyle: {},
          emphasis: {
            focus: "series"
          },
          data: [150, 232, 201, 154, 190, 330, 410]
        },
        {
          name: "直接访问",
          type: "line",
          stack: "总量",
          areaStyle: {},
          emphasis: {
            focus: "series"
          },
          data: [320, 332, 301, 334, 390, 330, 320]
        },
        {
          name: "搜索引擎",
          type: "line",
          stack: "总量",
          label: {
            show: true,
            position: "top"
          },
          areaStyle: {},
          emphasis: {
            focus: "series"
          },
          data: [820, 932, 901, 934, 1290, 1330, 1320]
        }
      ]
    });

    return { pieOption, barOption, areaOption };
  }
});
const AllStudent = {
  "languages": [
    {
      "language": "英语",
      "level": "四级",
      "length": 3
    },
    {
      "language": "英语",
      "level": "六级",
      "length": 5
    },
  ],
  "cities": [
    {
      "name": "北京",
      "length": 3
    },
    {
      "name": "大连",
      "length": 3
    },
    {
      "name": "上海",
      "length": 3
    },
  ],
  "internships": [
    {
      "company": "腾讯",
      "experience": "游戏后端开发",
      "length": 3
    },
    {
      "company": "阿里",
      "experience": "游戏后端开发",
      "length": 3
    },
  ],
  "specialities": [
    {
      "name": "Java",
      "length": 5
    },
    {
      "name": "Python",
      "length": 7
    },
    {
      "name": "C++",
      "length": 10
    },
  ],
  "jobs": [
    {
      "name": "后端开发",
      "length": 5
    },
    {
      "name": "前端开发",
      "length": 7
    },
    {
      "name": "客户端开发",
      "length": 10
    },
  ]
}
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
  border-radius: 10px;
  padding: 20px;
  background: #fff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart {
  height: 400px;
}
</style>