package com.resume.student.controller;

import com.resume.model.dto.ResumeUploadVO;
import com.resume.model.dto.StudentCompleteInfoDTO;
import com.resume.model.dto.StudentDetailDTO;
import com.resume.model.result.Result;
import com.resume.student.service.StudentInfoService;
import com.resume.utils.JwtUtil;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.multipart.MultipartFile;

/**
 * 学生端控制器
 */
@RestController
@RequestMapping("/stu")
@Tag(name = "学生端")
public class StudentController {
    
    @Autowired
    private StudentInfoService studentInfoService;
    private StudentController StudentInfoMapper;

    /**
     * 完善学生信息
     * @param infoDTO 完善信息DTO
     * @param request HTTP请求
     * @return 更新后的学生详细信息
     */
    @PutMapping("/completeInfo")
    @Operation(summary = "完善学生信息")
    public Result<StudentDetailDTO> completeInfo(@RequestBody StudentCompleteInfoDTO infoDTO, HttpServletRequest request) {
        try {
            // 从Token中获取用户ID
            Integer userId = (Integer) request.getAttribute("userId");
            
            StudentDetailDTO result = studentInfoService.completeInfo(userId, infoDTO);
            return Result.success(result);
        } catch (RuntimeException e) {
            String message = e.getMessage();
            
            // 如果消息为空，使用异常类名
            if (message == null || message.trim().isEmpty()) {
                e.printStackTrace(); // 打印完整堆栈以便调试
                return Result.error(500, "服务器内部错误: " + e.getClass().getSimpleName());
            }
            
            if (message.contains(":")) {
                String[] parts = message.split(":", 2);
                try {
                    int code = Integer.parseInt(parts[0]);
                    String msg = parts[1];
                    return Result.error(code, msg);
                } catch (NumberFormatException ex) {
                    // 如果无法解析错误码，返回原始消息
                    return Result.error(500, message);
                }
            }
            return Result.error(500, "服务器内部错误: " + message);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(500, "服务器内部错误: " + (e.getMessage() != null ? e.getMessage() : e.getClass().getSimpleName()));
        }
    }
    
    /**
     * 上传简历
     * @param file 简历文件
     * @param request HTTP请求
     * @return 上传结果
     */
    @PostMapping("/uploadResume")
    @Operation(summary = "上传简历")
    public Result<ResumeUploadVO> uploadResume(@RequestParam("resumeFile") MultipartFile file, HttpServletRequest request) {
        try {
            // 从Token中获取用户ID
            Integer userId = (Integer) request.getAttribute("userId");
            
            ResumeUploadVO result = studentInfoService.uploadResume(userId, file);
            return new Result<>(201, "简历上传成功", result);
        } catch (RuntimeException e) {
            String message = e.getMessage();
            
            // 如果消息为空，使用异常类名
            if (message == null || message.trim().isEmpty()) {
                e.printStackTrace(); // 打印完整堆栈以便调试
                return Result.error(500, "服务器内部错误: " + e.getClass().getSimpleName());
            }
            
            if (message.contains(":")) {
                String[] parts = message.split(":", 2);
                try {
                    int code = Integer.parseInt(parts[0]);
                    String msg = parts[1];
                    return Result.error(code, msg);
                } catch (NumberFormatException ex) {
                    // 如果无法解析错误码，返回原始消息
                    return Result.error(500, message);
                }
            }
            return Result.error(500, "服务器内部错误: " + message);
        } catch (Exception e) {
            e.printStackTrace();
            return Result.error(500, "服务器内部错误: " + (e.getMessage() != null ? e.getMessage() : e.getClass().getSimpleName()));
        }
    }


    @GetMapping("/statistics")
    public Result getAllStudentStatistics() {
        return Result.success(studentInfoService.getAllStudentStatistics());  // ✅ 正确
    }
}
