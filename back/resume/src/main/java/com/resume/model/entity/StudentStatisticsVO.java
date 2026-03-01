package com.resume.model.entity;

import lombok.Data;

import java.util.List;

@Data
public class StudentStatisticsVO {

    // 语言统计
    private List<LanguageStat> languages;

    // 城市统计
    private List<CityStat> cities;

    // 实习统计
    private List<InternshipStat> internships;

    // 技能统计
    private List<SpecialityStat> specialities;

    // 职位统计
    private List<JobStat> jobs;

    // 内部类定义
    @Data
    public static class LanguageStat {
        private Integer value;
        private String language;
        private String level;

    }

    @Data
    public static class CityStat {
        private Integer value;
        private String name;

    }

    @Data
    public static class InternshipStat {
        private Integer value;
        private String company;
        private String experience;

    }

    @Data
    public static class SpecialityStat {
        private Integer value;
        private String name;

    }

    @Data
    public static class JobStat {
        private Integer value;
        private String name;

    }
}