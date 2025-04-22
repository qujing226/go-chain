package main

import (
	"encoding/base64"
	"fmt"
	chaindid "github.com/qujing226/blockchain/did"
	"github.com/qujing226/blockchain/wallet"
	"github.com/stretchr/testify/require"
	"testing"
)

// 此过程是Alice向Bob发起一个私密的请求
//
//   1： Alice 从链方存储的DID doc中获取到 Bob 的KEM加密公钥
//	 2： Alice 获取通过公钥生成 共享密钥 和密文
//	 3： Bob 通过私钥解出 共享密钥 的明文

// Bob 的Kem钱包地址： iFSX4x2k9WnFirR6YXnAj54AWeJ7UXNyzW

func Aes256EncryptionAndDecrypt(t *testing.T) {
	testcases := []struct {
		name             string
		text             string
		encapsulationKey string
	}{
		{
			name:             "test",
			text:             "",
			encapsulationKey: "",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {

		})
	}

}

func TestEncryptWithKEMEncapsulationKey(t *testing.T) {
	testcases := []struct {
		name             string
		encapsulationKey string
	}{
		{
			name:             "test",
			encapsulationKey: "LAtS3tlfw9hi_4t2oHyUv8kVFQSlOOgfTvQAMtEk1eiAqISqSDk0UrZBfzVsGUJuvQeua7iOuAgliSwqloNwMNd8roW-FMBngpImsHlvBKWQ8XJafmCMTEwA8obGmro7z7YL1TkCOmhOH3u29yZq5NxekuMPNqDGbyDOdcwmXhw78mMj1OxQdsbCSLCO6vqKxfU6vqCnSMiLxuw5zphOYKsYQSFwu5RNknqWVomfJ7rCQYQ1v_PNwtCpjLF6GetjMgSJkiZb3gMn72RUTmYZlXXBbfEeB2ifYAx1G8RD_GWO79ON45xcOPSDhEo3FoA9NVqRfouUhzh5JrSATRPM0zcQZqEVMcu_iTSaiqGHSyqCAjzGdrCkGwDCdaaZT8c9MyJjywJ35JRHj2cTX-GAQjKuNFVNlLiOuIDFblejZfaXtWRHswly2XPD2RcfROeAwbdFDSRdiUx6fpLKD2F-8eixUSRahbEqGJINwdCcyfuWMlw5VtA6uhoLq-kngVWwQ1E1QVishMgM0RokAoYG3qqpWrli-BgYLDqlU2mPniR7QATHYdIe9HY5kaQbnzM6NlbE7vuk1EJWLfLJCbVD14KDNuFI52FFMKXL-yW0TIN-Ksg25fzKATW-NuTGfrqFw3omjqFIiDeii1yRqgCbMoYsIYeAeoxLj8uhXzoma_xXlrE-rzrGtgguuJCaLDG-uxNfs2MEaqYKPExstjcQuQcLlngbXSIQ_qE7kkDLKyAi7yaRUqdNobfEy2xbnvtWj9wnDkoaiRd-_TKwaLl0U0uFgMxdVvvEc0cRsHOIC3UTg9J2qzFTW0obDnOS5ZwDP6sPN-qPMPaSQBg384O2UxEZWCwNTTx52LZnUMlg5jwgN5TGW9lR64c3Kkenqzx3uTIjT1Mi8GebPckJUKBPw_QZXKV52QeFK1lRwDy6W2Jhy9RKbdizXiKnYmzK34hLiYuxL4BxP7o_7RhcSQBi5Hy5KvbLA7SEnINyFJDEAXKqyOlJdccPUkpQv-Cp1-I6rki6A9lgvgiXXHiF_ridY5qjkjo1wWkBWnNZYbhk2GcJ20CdUekWo2WiRaun_UC5NlkVzTtpXvWIqDV_y_hU2UjBZzIuzdJZK4U98mZy7knKwswYavEpXfGH9jlAoiQhH7Qe3WcYLik6bjqg4Mdm63wuVZlp34BtjWcsQBiYs0TIQnt1Z8EbejNfL-Ihrmq3l0pE_QQ6ypVqFnI8JQCGmYgmF4o7REtcEOrM_vVArvwqVTIKsbUPhSu6vDxrCDQB9HO8mgwB8nJ8sWsf49uOSINoEMlylgGBHbuo6ec3S8yl2ObLHIQGgZhIO7yIMqWkrGGdoKwGxFIn3tgK-JqyvhdRgvkZVvuqgHHPEuEQ1EBipVBxKEXG13bLYbC6MnnH8CN_3yhbF3uKf9CeZ0HKWTeifwSemRIEwOCyGYciFIJEpNlaTWQEu4gz7NUK7fKgCEgXoiQ-_mORZiN1n5aqCyenpqwW2hufmhehKFGPg7CYxeYOj_w8lMOs2Pgcp5qTyXPk2tUgLf6VFXhgGkdQq98EDga5scRk92zZlKETdWU",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			en, err := base64.RawURLEncoding.DecodeString(tc.encapsulationKey)
			require.NoError(t, err)
			sharedSecretReceiver, ciphertext, err := chaindid.EncryptWithKEM([1184]byte(en))
			require.NoError(t, err)
			fmt.Println("sharedSecretReceiver: ", sharedSecretReceiver)
			fmt.Println("ciphertext: ", base64.RawURLEncoding.EncodeToString(ciphertext[:]))
		})
	}
}

func TestDecryptWithKEMDecapsulationKey(t *testing.T) {
	testcases := []struct {
		name       string
		ciphertext string
	}{
		{
			name:       "test",
			ciphertext: "rDqhLXAEO5bgHK8CJg53euTewixi1EuUIHMnmY5w10k-afodRbOpnXHpHUPwx3RmtAp0qpq5sO6mxMe3BKFAcWKW1DK4bGim9oIXOrTW_7uHQE8oyLwHRaDGaL2b20RVsRO8u5XAGDnz3s_iLIrsqYkligN3T6efB3md3ut9AgoEAF8JdCtiLqHLgat_H9EcQQ6giZyf1tGA2lXVtrcSBHdOPkDtBiLl-S-AjH4PQUgv_pWbLqYjQVYFKTWm2pLuwlmVlb9aWoccqNKu_ENP4zSe17cj2b_OG-QUKriJn_0Cmbm9TbMclY89VfFF-mC9iKZA7xdxZjlrtdf_tOfIb9u_e7tTlWIYoZHST_0aQVxYo9qPYto90fYBn4nB_whtGcFiO1DXEW8x19PXfJsqL2dC-ru5MvfyzGmAWvyRVlsrIwFgUfX0oME5vTJLejDs0f53vGmo5mbTXlLEsnRktaLapF06HYxrJ5deJVqOKK0o29MKnCG3W_8GcwU7R40Lnj6CsWptiT-qJREaZTcGidBWOmuXiBploI5K8FjMGQvwk3lQVyE7jbuH60acQYjtafrAODZ1LzxOKHShpFRlHMzNgDzX3su0JYfsgrxlIhhofLeZrTYgP2Qm04oQqBN7UXU0lYu7l9RlQuQlwmf0wHMlGFszNiEjGSfynYheAJ1mXuctOhDgqLahL7i8aWewYTVWiNksbfIo0MQAe8uctcfe0siDD6GIO2iQSs0Jk22AzrKbT3Sli4Alaf7JrRFGm_n7iLRV4JG_Ba7t_b089bCgEfQaRGFBe_rwv97ImcUndx0kKkxG5-Tk7pkUqpjD92pUIfjF6P2pL_ejoedw5L6_GbrRBOFFC5GfJxi8WsCVpob2-ednUji6c1pVdkZ0x6_pwaJPIe3x3XR477mWgrL0Yv8vxOrLnpgZKQgawh-L5Hl7N4mGugV8vHnjR8yNZW9SlYjAlqLLcszrVrMLisYNR1sltxrrQQPOGoWKtszqG0VDxMHLzgYxkvRuhTbypGiqZa-l1T68_G1U0UzyTUyhcGs3hRp1BI_r7RZWQlrsmuGxK1pe1LAqcWjeSNsmJVy8Wi3_ZcH77o8S48T4eNQsFsPTcyPlAx-plA2a_feGuChSiteOi64vb1gexqRo5dnuApc9PmnOi6Gw_C04TTpeie2jUY_20rPVLSaSHAPxG1MPBmRtMTvWai9_JH7hW60LONZSjd0uHaXAh_kybmXigt4cDEfA9VMpNAou8tmTMVvm78WV34AQcbsj8dNxpGTqZYyZTuUjEz9Ycg7k_J-NGp5_axUHaO9n3hqh-Hu2w9-mjSOQ-_dvWaFfLcV8ocx9p1948qfU0Zxb3_R4pPfLoUUGppd50vLN-hKovA6BnyoWn9KhFurF6DXhaEpfZTAMaSXuCWXX23r2ogv6OZURtYpJMl7vcEnfWmjZ87I",
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ciphertext, err := base64.RawURLEncoding.DecodeString(tc.ciphertext)
			require.NoError(t, err)

			kws, err := wallet.NewKemWallets()
			require.NoError(t, err)
			kw := kws.GetWallet("iFSX4x2k9WnFirR6YXnAj54AWeJ7UXNyzW")
			sharedSecretKey, err := chaindid.DecryptWithKEM(kw.DecapsulationKey, [1088]byte(ciphertext))
			require.NoError(t, err)
			fmt.Println("sharedSecretKey: ", sharedSecretKey)
		})
	}
}
