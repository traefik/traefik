package sacloud

// AllowArchiveSizes 作成できるアーカイブのサイズ
func AllowArchiveSizes() []string {
	return []string{"20", "40", "60", "80", "100", "250", "500", "750", "1024"}
}

// Archive アーカイブ
type Archive struct {
	*Resource             // ID
	propAvailability      // 有功状態
	propName              // 名称
	propDescription       // 説明
	propSizeMB            // サイズ(MB単位)
	propMigratedMB        // コピー済みデータサイズ(MB単位)
	propScope             // スコープ
	propCopySource        // コピー元情報
	propServiceClass      // サービスクラス
	propPlanID            // プランID
	propJobStatus         // マイグレーションジョブステータス
	propOriginalArchiveID // オリジナルアーカイブID
	propStorage           // ストレージ
	propBundleInfo        // バンドル情報
	propTags              // タグ
	propIcon              // アイコン
	propCreatedAt         // 作成日時
}
