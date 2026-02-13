<template>
  <div class="charts-container">
    <div style="display: flex;flex-direction: row;gap:15px">
      <div class="chart-wrapper" style="width:48%">
        <v-chart class="chart" :option="radarOption" />
      </div>
      <div class="chart-wrapper" style="width:48%">
        <v-chart class="chart" :option="pieOptionJobs" />
      </div>
    </div>
    <div class="chart-wrapper">
      <v-chart class="chart" :option="barOption" />
    </div>
    <div class="chart-wrapper">
      <v-chart class="chart" :option="pieOptionSpecialities" />
    </div>
    <div class="chart-wrapper">
      <v-chart class="chart" :option="areaOption" />
    </div>
  </div>
</template>

<script>
import { use } from "echarts/core";
import { CanvasRenderer } from "echarts/renderers";
import { PieChart, BarChart, LineChart, RadarChart } from "echarts/charts";
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
  RadarChart,
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
    const a = {
      "languages": [
        {
          "value": 130,
          "language": "英语",
          "level": "四级"
        },
        {
          "value": 50,
          "language": "韩语",
          "level": "六级"
        },
        {
          "value": 56,
          "language": "汉语",
          "level": "六级"
        },
        {
          "value": 136,
          "language": "法语",
          "level": "六级"
        },
        {
          "value": 53,
          "language": "日语",
          "level": "六级"
        },
      ],
      "cities": [
        {
          "value": 35,
          "name": "北京"
        },
        {
          "value": 39,
          "name": "大连"
        },
        {
          "value": 135,
          "name": "上海"
        },
        {
          "value": 79,
          "name": "沈阳"
        },
        {
          "value": 45,
          "name": "杭州"
        },
      ],
      "internships": [
        {
          "value": 101,
          "company": "腾讯",
          "experience": "游戏后端开发"
        },
        {
          "value": 79,
          "company": "阿里",
          "experience": "游戏后端开发"
        },
        {
          "value": 232,
          "company": "腾讯天美",
          "experience": "游戏后端开发"
        },
        {
          "value": 382,
          "company": "阿里淘天",
          "experience": "游戏后端开发"
        },
      ],
      "specialities": [
        {
          "value": 500,
          "name": "Java"
        },
        {
          "value": 370,
          "name": "Python"
        },
        {
          "value": 100,
          "name": "C++"
        },
        {
          "value": 380,
          "name": "SpringBoot"
        },
        {
          "value": 200,
          "name": "Vue2"
        },
        {
          "value": 200,
          "name": "Vue3"
        },
      ],
      "jobs": [
        {
          "value": 56,
          "name": "后端开发"
        },
        {
          "value": 179,
          "name": "前端开发"
        },
        {
          "value": 210,
          "name": "客户端开发"
        },
        {
          "value": 170,
          "name": "大模型算法"
        },
      ]
    }
    const radarOption = ref({
      title: {
        text: "应届生技能水平雷达图",
        left: "center"
      },
      tooltip: {},
      radar: {
        indicator: a.specialities.map(item=>({
          name:item.name,
          value:item.value
        }))
      },
      series: [{
        type: "radar",
        data: [{
          value: a.specialities.map(item => item.value),
          name: "技能水平",
          areaStyle: {}
        }]
      }]
    });
    const pieOptionSpecialities = ref({
      title: {
        text: "应届生实习经历",
        left: "center"
      },
      tooltip: {
        trigger: "item",
        formatter: "{a} <br/>{b} : {c} ({d}%)"
      },
      legend: {
        orient: "vertical",
        left: "left",
        data: a.internships.map(item => item.company)
      },
      series: [
        {
          name: "Traffic Sources",
          type: "pie",
          radius: "55%",
          center: ["50%", "60%"],
          data: a.internships.map(item => ({
            value: item.value,
            name: item.company
          })),
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: "rgba(0,0, 0, 0.5)"
            }
          }
        }
      ]
    });
    const pieOptionJobs = ref({
      title: {
        text: "应届生意向工作分布",
        left: "center"
      },
      tooltip: {
        trigger: "item",
        formatter: "{a} <br/>{b} : {c} ({d}%)"
      },
      legend: {
        orient: "vertical",
        left: "left",
        data: a.jobs.map(item => item.name)
      },
      series: [
        {
          name: "Traffic Sources",
          type: "pie",
          radius: "55%",
          center: ["50%", "60%"],
          data: a.jobs.map(item => ({
            value: item.value,
            name: item.name
          })),
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
        text: "应届生意向就业城市",
        left: "center"
      },
      tooltip: {
        trigger: "axis",
        axisPointer: {
          type: "shadow"
        }
      },
      legend: {
        data: ["意向人数"],
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
        data: a.cities.map(item => item.name)
      },
      yAxis: {
        type: "value"
      },
      series: [
        {
          name: "意向人数",
          type: "bar",
          data: a.cities.map(item => item.value),
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
    return { pieOptionSpecialities, pieOptionJobs, barOption, areaOption, radarOption };
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
  width: 97%;
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