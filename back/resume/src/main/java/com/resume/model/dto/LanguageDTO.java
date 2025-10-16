package com.resume.model.dto;

import lombok.Data;
import java.util.Date;

@Data
public class LanguageDTO {
    private String language;
    private String level;
    private String certNo;
    private Date issueDate;
}