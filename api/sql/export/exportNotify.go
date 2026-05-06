package export

import (
	"cmdb/models/sql"
	"cmdb/pkg/middleware"
	"cmdb/pkg/notice"
	"fmt"
)

// SendExportNotification 发送导出完成通知邮件
// success: true-成功, false-失败
// rowCount: 导出的记录条数(仅成功时有效)
// errorMsg: 失败原因(仅失败时有效)
// downloadURL: 下载链接(仅成功时有效)
// expireAt: 链接过期时间(仅成功时有效)
func SendExportNotification(export *sql.SQLExport, success bool, rowCount int64, errorMsg string, downloadURL string, expireAt string) error {
	// 获取项目中文名称
	projectName := sql.GetProjectName(export.Project)
	if projectName == "" {
		projectName = export.Project
	}

	if success {
		// 成功时发送给提交人,抄送审核人
		return sendSuccessNotification(export, projectName, rowCount, downloadURL, expireAt)
	} else {
		// 失败时发送给执行人
		return sendFailureNotification(export, projectName, errorMsg)
	}
}

// sendSuccessNotification 发送成功通知
func sendSuccessNotification(export *sql.SQLExport, projectName string, rowCount int64, downloadURL string, expireAt string) error {
	// 获取提交人邮箱
	submitterEmail := sql.GetUserEmail(export.SubmitterID)
	if submitterEmail == "" {
		middleware.Logger.Errorf("无法获取提交人 %d 的邮箱", export.SubmitterID)
		return fmt.Errorf("无法获取提交人邮箱")
	}

	// 获取审批人邮箱(抄送)
	approverEmail := sql.GetUserEmail(export.ApplyID)
	if approverEmail == "" {
		middleware.Logger.Warnf("无法获取审批人 %d 的邮箱,将不抄送", export.ApplyID)
	}

	// 构建邮件主题
	subject := fmt.Sprintf("【SQL导出完成】%s - 导出成功", projectName)

	// 构建下载链接部分的HTML
	downloadSection := ""
	if downloadURL != "" {
		downloadSection = fmt.Sprintf(`
            <div class="info-row">
                <span class="label">下载链接:</span>
                <span class="value"><a href="%s" style="color: #4CAF50; text-decoration: none;">点击下载</a></span>
            </div>
            <div class="info-row">
                <span class="label">链接有效期:</span>
                <span class="value">%s (10分钟)</span>
            </div>`, downloadURL, expireAt)
	}

	// 构建邮件正文
	body := fmt.Sprintf(`
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 15px; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 20px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .info-row { margin: 10px 0; }
        .label { font-weight: bold; color: #555; }
        .value { color: #333; }
        .success { color: #4CAF50; font-weight: bold; }
        .footer { margin-top: 20px; padding-top: 15px; border-top: 1px solid #ddd; font-size: 12px; color: #777; }
        a { color: #4CAF50; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2 style="margin: 0;">SQL 导出完成通知</h2>
        </div>
        <div class="content">
            <div class="info-row">
                <span class="label">项目名称:</span>
                <span class="value">%s</span>
            </div>
            <div class="info-row">
                <span class="label">提交说明:</span>
                <span class="value">%s</span>
            </div>
            <div class="info-row">
                <span class="label">导出结果:</span>
                <span class="success">✓ 导出成功</span>
            </div>
            <div class="info-row">
                <span class="label">记录条数:</span>
                <span class="value">%d 条</span>
            </div>
            %s
            <div class="footer">
                <p>此邮件为系统自动发送,请勿直接回复。</p>
                <p>下载链接有效期为10分钟,过期后请登录系统重新生成。</p>
            </div>
        </div>
    </div>
</body>
</html>
`, projectName, export.SubmitterRemark, rowCount, downloadSection)

	// 准备收件人和抄送人列表
	toList := []string{submitterEmail}
	ccList := []string{}
	if approverEmail != "" {
		ccList = append(ccList, approverEmail)
	}

	// 发送邮件
	emailInfo := &notice.EmailInfo{
		To:       toList,
		Cc:       ccList,
		Subject:  subject,
		Body:     body,
		BodyType: "text/html",
	}

	if err := notice.SendEmail(emailInfo); err != nil {
		middleware.Logger.Errorf("发送成功通知邮件失败: %v", err)
		return err
	}

	middleware.Logger.Infof("SQL导出成功通知已发送 - 收件人: %s, 抄送: %v, 流程ID: %d",
		submitterEmail, ccList, export.ID)
	return nil
}

// sendFailureNotification 发送失败通知
func sendFailureNotification(export *sql.SQLExport, projectName string, errorMsg string) error {
	// 获取执行人邮箱
	executorEmail := sql.GetUserEmail(export.ExecutorID)
	if executorEmail == "" {
		middleware.Logger.Errorf("无法获取执行人 %d 的邮箱", export.ExecutorID)
		return fmt.Errorf("无法获取执行人邮箱")
	}

	// 构建邮件主题
	subject := fmt.Sprintf("【SQL导出失败】%s - 导出失败", projectName)

	// 构建邮件正文
	body := fmt.Sprintf(`
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f44336; color: white; padding: 15px; border-radius: 5px 5px 0 0; }
        .content { background-color: #f9f9f9; padding: 20px; border: 1px solid #ddd; border-radius: 0 0 5px 5px; }
        .info-row { margin: 10px 0; }
        .label { font-weight: bold; color: #555; }
        .value { color: #333; }
        .error { color: #f44336; font-weight: bold; }
        .error-detail { background-color: #ffebee; padding: 10px; border-left: 3px solid #f44336; margin: 10px 0; }
        .footer { margin-top: 20px; padding-top: 15px; border-top: 1px solid #ddd; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2 style="margin: 0;">SQL 导出失败通知</h2>
        </div>
        <div class="content">
            <div class="info-row">
                <span class="label">项目名称:</span>
                <span class="value">%s</span>
            </div>
            <div class="info-row">
                <span class="label">提交说明:</span>
                <span class="value">%s</span>
            </div>
            <div class="info-row">
                <span class="label">导出结果:</span>
                <span class="error">✗ 导出失败</span>
            </div>
            <div class="error-detail">
                <div class="label">失败原因:</div>
                <div style="margin-top: 5px;">%s</div>
            </div>
            <div class="footer">
                <p>此邮件为系统自动发送,请勿直接回复。</p>
                <p>请检查失败原因后重新执行导出操作。</p>
            </div>
        </div>
    </div>
</body>
</html>
`, projectName, export.SubmitterRemark, errorMsg)

	// 准备收件人列表(仅执行人,无抄送)
	toList := []string{executorEmail}

	// 发送邮件
	emailInfo := &notice.EmailInfo{
		To:       toList,
		Subject:  subject,
		Body:     body,
		BodyType: "text/html",
	}

	if err := notice.SendEmail(emailInfo); err != nil {
		middleware.Logger.Errorf("发送失败通知邮件失败: %v", err)
		return err
	}

	middleware.Logger.Infof("SQL导出失败通知已发送 - 收件人: %s, 流程ID: %d",
		executorEmail, export.ID)
	return nil
}
