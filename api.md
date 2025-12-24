# 图片分析模块 API 接口文档

## 概述

本文档描述了图片分析模块的所有API接口，包括图片上传、分析任务管理、结果查询等功能。

## 基础信息

- **基础URL**: `/api`
- **认证方式**: Bearer Token
- **数据格式**: JSON
- **字符编码**: UTF-8

## 通用响应格式

### 成功响应
```json
{
  "success": true,
  "message": "操作成功",
  "data": {
    // 具体数据
  }
}
```

### 错误响应
```json
{
  "success": false,
  "message": "错误信息",
  "error": "具体错误描述"
}
```

## 1. 图片上传接口

### 接口信息
- **接口地址**: `POST /api/upload/image`
- **功能描述**: 上传课堂图片文件
- **请求方式**: multipart/form-data

### 请求头
```
Authorization: Bearer {teacher_token}
Content-Type: multipart/form-data
```

### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| file | File | 是 | 图片文件 |
| type | String | 否 | 文件类型标识，固定值："image_analysis" |

### 响应示例

#### 成功响应
```json
{
  "success": true,
  "message": "上传成功",
  "data": {
    "imageId": "img_1234567890",
    "url": "https://example.com/uploads/image_1234567890.jpg",
    "filename": "original_filename.jpg",
    "size": 1024000,
    "uploadTime": "2024-01-01T12:00:00Z"
  }
}
```

#### 错误响应
```json
{
  "success": false,
  "message": "文件格式不支持",
  "error": "只支持 JPG、PNG、BMP 格式的图片"
}
```

### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| imageId | String | 图片唯一标识符 |
| url | String | 图片访问URL |
| filename | String | 原始文件名 |
| size | Number | 文件大小(字节) |
| uploadTime | String | 上传时间(ISO 8601格式) |

## 2. 开始图片分析接口

### 接口信息
- **接口地址**: `POST /api/teacher/image/analyze`
- **功能描述**: 创建图片分析任务
- **请求方式**: multipart/form-data

### 请求头
```
Authorization: Bearer {teacher_token}
Content-Type: multipart/form-data
```

### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| files[] | File[] | 是 | 图片文件数组 |
| title | String | 是 | 课程名称 |
| className | String | 是 | 班级名称 |
| studentCount | Number | 是 | 预计学生数量(1-100) |
| analysisMode | String | 是 | 分析模式："facial"\|"posture"\|"comprehensive" |
| description | String | 否 | 分析说明 |

### 分析模式说明
- `facial`: 面部表情分析
- `posture`: 姿态分析
- `comprehensive`: 综合分析

### 响应示例

#### 成功响应
```json
{
  "success": true,
  "message": "分析任务已创建",
  "data": {
    "taskId": "task_1234567890",
    "status": "processing",
    "estimatedTime": 30
  }
}
```

#### 错误响应
```json
{
  "success": false,
  "message": "参数错误",
  "error": "课程名称不能为空"
}
```

### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| taskId | String | 分析任务唯一标识符 |
| status | String | 任务状态："pending"\|"processing"\|"completed"\|"failed" |
| estimatedTime | Number | 预计完成时间(秒) |

## 3. 分析进度查询接口

### 接口信息
- **接口地址**: `GET /api/teacher/image/analysis-status`
- **功能描述**: 查询分析任务进度
- **请求方式**: GET

### 请求头
```
Authorization: Bearer {teacher_token}
```

### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| taskId | String | 是 | 任务ID |

### 响应示例

#### 成功响应
```json
{
  "success": true,
  "data": {
    "taskId": "task_1234567890",
    "status": "processing",
    "progress": 75,
    "currentStep": "人脸检测",
    "estimatedTime": 10
  }
}
```

#### 错误响应
```json
{
  "success": false,
  "message": "任务不存在",
  "error": "taskId: task_1234567890 不存在"
}
```

### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| taskId | String | 任务ID |
| status | String | 任务状态 |
| progress | Number | 进度百分比(0-100) |
| currentStep | String | 当前处理步骤 |
| estimatedTime | Number | 剩余时间(秒) |

### 任务状态说明
- `pending`: 等待中
- `processing`: 处理中
- `completed`: 已完成
- `failed`: 失败

## 4. 获取分析结果接口

### 接口信息
- **接口地址**: `GET /api/teacher/image/analysis-result`
- **功能描述**: 获取分析结果详情
- **请求方式**: GET

### 请求头
```
Authorization: Bearer {teacher_token}
```

### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| taskId | String | 是 | 任务ID |

### 响应示例

#### 成功响应
```json
{
  "success": true,
  "data": {
    "id": "analysis_1234567890",
    "title": "数学课",
    "className": "高三一班",
    "analysisTime": "2024-01-01T12:00:00Z",
    "imageCount": 3,
    "avgFocus": 85.5,
    "focusedStudents": 18,
    "distractedStudents": 7,
    "totalStudents": 25,
    "distributionData": [
      {
        "name": "高度专注(90-100%)",
        "value": 8,
        "color": "#67c23a"
      },
      {
        "name": "良好专注(80-89%)",
        "value": 10,
        "color": "#409eff"
      },
      {
        "name": "一般专注(70-79%)",
        "value": 5,
        "color": "#e6a23c"
      },
      {
        "name": "注意力分散(<70%)",
        "value": 2,
        "color": "#f56c6c"
      }
    ],
    "imageResults": [
      {
        "id": "img_1",
        "url": "https://example.com/uploads/image_1.jpg",
        "filename": "classroom_1.jpg",
        "faces": [
          {
            "face_index": 1,
            "focus_score": 0.92,
            "confidence": 0.95,
            "class_num": 1,
            "bbox": {
              "x1": 100,
              "y1": 150,
              "x2": 200,
              "y2": 250
            },
            "landmarks": [
              {
                "x": 120,
                "y": 160
              },
              {
                "x": 180,
                "y": 160
              }
            ]
          }
        ]
      }
    ]
  }
}
```

### 字段说明

#### 基础信息
| 字段名 | 类型 | 说明 |
|--------|------|------|
| id | String | 分析记录ID |
| title | String | 课程名称 |
| className | String | 班级名称 |
| analysisTime | String | 分析时间(ISO 8601格式) |
| imageCount | Number | 分析图片数量 |

#### 统计摘要
| 字段名 | 类型 | 说明 |
|--------|------|------|
| avgFocus | Number | 平均专注度(%) |
| focusedStudents | Number | 专注学生数 |
| distractedStudents | Number | 分心学生数 |
| totalStudents | Number | 总学生数 |

#### 分布数据
| 字段名 | 类型 | 说明 |
|--------|------|------|
| distributionData | Array | 专注度分布数据(用于饼图) |
| name | String | 分布区间名称 |
| value | Number | 学生数量 |
| color | String | 显示颜色 |

#### 图片结果
| 字段名 | 类型 | 说明 |
|--------|------|------|
| imageResults | Array | 图片分析结果数组 |
| id | String | 图片ID |
| url | String | 图片URL |
| filename | String | 文件名 |
| faces | Array | 检测到的人脸数组 |

#### 人脸数据
| 字段名 | 类型 | 说明 |
|--------|------|------|
| face_index | Number | 人脸序号 |
| focus_score | Number | 专注度分数(0-1) |
| confidence | Number | 检测置信度(0-1) |
| class_num | Number | 类别编号 |
| bbox | Object | 人脸边界框 |
| landmarks | Array | 人脸关键点数组 |

## 5. 获取历史分析记录接口

### 接口信息
- **接口地址**: `GET /api/teacher/image/history`
- **功能描述**: 获取历史分析记录列表
- **请求方式**: GET

### 请求头
```
Authorization: Bearer {teacher_token}
```

### 请求参数
| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| page | Number | 否 | 页码，默认1 |
| pageSize | Number | 否 | 每页数量，默认10 |
| startDate | String | 否 | 开始日期，格式：YYYY-MM-DD |
| endDate | String | 否 | 结束日期，格式：YYYY-MM-DD |

### 响应示例

#### 成功响应
```json
{
  "success": true,
  "data": {
    "total": 25,
    "page": 1,
    "pageSize": 10,
    "list": [
      {
        "id": "analysis_1234567890",
        "title": "数学课",
        "className": "高三一班",
        "analysisTime": "2024-01-01T12:00:00Z",
        "imageCount": 3,
        "avgFocus": 85.5,
        "focusedStudents": 18,
        "distractedStudents": 7,
        "status": "completed"
      }
    ]
  }
}
```

### 字段说明
| 字段名 | 类型 | 说明 |
|--------|------|------|
| total | Number | 总记录数 |
| page | Number | 当前页码 |
| pageSize | Number | 每页数量 |
| list | Array | 记录列表 |
| id | String | 分析记录ID |
| title | String | 课程名称 |
| className | String | 班级名称 |
| analysisTime | String | 分析时间 |
| imageCount | Number | 图片数量 |
| avgFocus | Number | 平均专注度 |
| focusedStudents | Number | 专注学生数 |
| distractedStudents | Number | 分心学生数 |
| status | String | 分析状态 |

## 错误码说明

| HTTP状态码 | 错误码 | 说明 |
|-----------|--------|------|
| 400 | INVALID_PARAMETER | 参数错误 |
| 401 | UNAUTHORIZED | 未授权，token无效 |
| 403 | FORBIDDEN | 权限不足 |
| 404 | NOT_FOUND | 资源不存在 |
| 413 | FILE_TOO_LARGE | 文件过大 |
| 415 | UNSUPPORTED_FORMAT | 不支持的文件格式 |
| 429 | TOO_MANY_REQUESTS | 请求过于频繁 |
| 500 | INTERNAL_ERROR | 服务器内部错误 |
| 503 | SERVICE_UNAVAILABLE | 服务不可用 |

## 业务限制

### 文件上传限制
- **支持格式**: JPG、PNG、BMP
- **单文件大小**: 最大10MB
- **批量上传**: 最多5张图片
- **文件命名**: 建议使用有意义的文件名

### 分析限制
- **并发任务**: 每个教师最多同时运行3个分析任务
- **图片分辨率**: 建议1920x1080或以上
- **人脸检测**: 每张图片最多检测50个人脸
- **分析时间**: 单张图片分析时间不超过30秒

### 数据保留
- **原始图片**: 保留30天
- **分析结果**: 保留1年
- **历史记录**: 保留2年

## 最佳实践

### 1. 图片质量要求
- 光线充足，避免过暗或过亮
- 学生面部清晰可见
- 避免严重遮挡
- 建议正面或侧面角度

### 2. 性能优化
- 使用异步处理避免阻塞
- 实现进度轮询机制
- 合理设置超时时间
- 实现任务队列管理

### 3. 错误处理
- 提供详细的错误信息
- 实现重试机制
- 记录错误日志
- 提供用户友好的提示

### 4. 安全考虑
- 验证文件类型和大小
- 检查文件内容安全性
- 限制上传频率
- 保护用户隐私数据

## 更新日志

| 版本 | 日期 | 更新内容 |
|------|------|----------|
| v1.0.0 | 2024-01-01 | 初始版本 |
| v1.1.0 | 2024-01-15 | 添加历史记录接口 |
| v1.2.0 | 2024-02-01 | 优化错误处理机制 |