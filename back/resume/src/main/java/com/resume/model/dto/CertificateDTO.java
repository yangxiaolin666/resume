package com.resume.model.dto;

import lombok.Data;
import java.util.Date;

@Data
public class CertificateDTO {
    private String certName;
    private String level;
    private String certNo;
    private Date issueDate;
}