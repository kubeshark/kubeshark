package oas

import (
	"testing"
)

func TestNegative(t *testing.T) {
	cases := []string{
		"",
		"{}",
		"0.0.29",
		"0.1",
		"1.0",
		"1.0.0",
		"2.1.73",
		"abTestV2",
		"actionText,setName,setAttribute,save,ignore,onEnd,getContext,end,get",
		"AddUserGroupLink",
		"advert-management.adBlockerMessage.html",
		"agents.author.1.json",
		"animated-gif",
		"b", // can be valid hexadecimal
		"big-danger-coronavirus-panic-greater-crisis",
		"breakout-box",
		"callback",
		"core.algorithm_execution.view",
		"core.devices.view",
		"data.json",
		"dialog.overlay.infinity.json",
		"domain-input",
		"embeddable", // lul, it's a valid HEX!
		"embeddable_blip",
		"E PLURIBUS UNUM",
		"etc",
		"eu-central-1a",
		"fcgi-bin",
		"footer.include.html",
		"fullHashes:find",
		"generate-feed",
		"GetAds",
		"GetCart",
		"GetUniversalVariableUser",
		"github-audit-exports",
		"g.js",
		"g.pixel",
		".html",
		"Hugo Michiels",
		"image.sbix",
		"index.html",
		"iPad",
		"Joanna Mazewski",
		"LibGit2Sharp",
		"Michael_Vaughan1.png",
		"New RSS feed has been generated",
		"nick-clegg",
		"opt-out",
		"pixel_details.html",
		"post.json",
		"profile-method-info",
		"project-id",
		"publisha.1.json",
		"publish_and_moderate",
		"Ronna McDaniel",
		"rtb-h",
		"runs",
		"sign-up",
		"some-uuid-maybe",
		"stable-4.0-version.json",
		"StartUpCheckout",
		"Steve Flunk",
		"sync_a9",
		"Ted Cruz",
		"test.png",
		"token",
		"ToList",
		"v2.1.3",
		"VersionCheck.php",
		"v Rusiji",
		"Walnut St",
		"web_widget",
		"zoom_in.cur",
		"xray",
		"web",
		"vipbets1",
		"trcc",
		"fbpixel",

		// TODO below
		// "tcfv2",
		// "Matt-cartoon-255x206px-small.png",
		// "TheTelegraph_portal_white-320-small.png",
		// "testdata-10kB.js",
	}

	for _, str := range cases {
		if IsGibberish(str) {
			t.Errorf("Mistakenly true: %s", str)
		}
	}
}

func TestPositive(t *testing.T) {
	cases := []string{
		"0a0d0174-b338-4520-a1c3-24f7e3d5ec50.html",
		"1024807212418223",
		"11ca096cbc224a67360493d44a9903",
		"1553183382779",
		"1554507871",
		"19180481",
		"203ef0f713abcebd8d62c35c0e3f12f87d71e5e4",
		"456795af-b48f-4a8d-9b37-3e932622c2f0",
		"601a2bdcc5b69137248ddbbf",
		"60fe9aaeaefe2400012df94f",
		"610bc3fd5a77a7fa25033fb0",
		"610bd0315a77a7fa25034368",
		"610bd0315a77a7fa25034368zh",
		"6120c057c7a97b03f6986f1b",
		"710a462e",
		"730970532670-compute@developer.gserviceaccount.com",
		"819db2242a648b305395537022523d65",
		"952bea17-3776-11ea-9341-42010a84012a",
		"a3226860758.html",
		"AAAA028295945",
		"arn-aws-ecs-eu-west-2-396248696294-cluster-london-01-ECSCluster-27iuIYva8nO4",
		"arn-aws-ecs-eu-west-2-396248696294-cluster-london-01-ECSCluster-27iuIYva8nO4", // ?
		"bnjksfd897345nl098asd53412kl98",
		"c738338322370b47a79251f7510dd",                        // prefixed hex
		"ci12NC01YzkyNTEzYzllMDRhLTAtYy5tb25pdG9yaW5nLmpzb24=", // long base64
		"css/login.0f48c49a34eb53ea4623.min.css",
		"d_fLLxlhzDilixeBEimaZ5",
		"e21f7112-3d3b-4632-9da3-a4af2e0e9166",
		"e8782afc112720300c049ff124434b79",
		"fb6cjraf9cejut2a",
		"i-0236530c66ed02200",
		"JEHJW4BKVFDRTMTUQLHKK5WVAU",
		"john.dow.1981@protonmail.com",
		"MDEyOk9yZ2FuaXphdGlvbjU3MzI0Nzk1",
		"MNUTGVFMGLEMFTBH0XSE5E02F6J2DS",
		"n63nd45qsj",
		"n9z9QGNiz",
		"NC4WTmcy",
		"proxy.3d2100fd7107262ecb55ce6847f01fa5.html",
		"QgAAAC6zw0qH2DJtnXe8Z7rUJP0FgAFKkOhcHdFWzL1ZYggtwBgiB3LSoele9o3ZqFh7iCBhHbVLAnMuJ0HF8hEw7UKecE6wd-MBXgeRMdubGydhAMZSmuUjRpqplML40bmrb8VjJKNZswD1Cg",
		"QgAAAC6zw0qH2DJtnXe8Z7rUJP0rG4sjLa_KVLlww5WEDJ__30J15en-K_6Y68jb_rU93e2TFY6fb0MYiQ1UrLNMQufqODHZUl39Lo6cXAOVOThjAMZSmuVH7n85JOYSCgzpvowMAVueGG0Xxg",
		"qwerqwerasdfqwer@protonmai.com",
		"r-ext-5579e00a95c90",
		"r-ext-5579e8b12f11e",
		"r-v4-5c92513c9e04a",
		"r-v4-5c92513c9e04a-0-c.monitoring.json",
		"segments-1563566437171.639994",
		"sp_ANQXRpqH_urn$3Auri$3Abase64$3A6698b0a3-97ad-52ce-8fc3-17d99e37a726",
		"sp_dxJTfx11_576742227280287872",
		"sp_NnUPB5wj_601a2bdcc5b69137248ddbbf",
		"sp_NxITuoE4_premiumchron-article-14302157_c_ryGQBs_r_yIWvwP",
		"t_52d94268-8810-4a7e-ba87-ffd657a6752f",
		"timeouts-1563566437171.639994",
		"u_YPF3GsGKMo02",

		"0000000000 65535 f",
		"0000000178 00000 n",
		"0-10000",
		"01526123,",
		"0,18168,183955,3,4,1151616,5663,731,223,5104,207,3204,10,1051,175,364,1435,4,60,576,241,383,246,5,1102",
		"05/10/2020",
		"14336456724940333",
		"fb6cjraf9cejut2a",
		"JEHJW4BKVFDRTMTUQLHKK5WVAU",

		// TODO
		/*
			"0,20",
			"0.001",
			"YISAtiX1",
			"Fxvd1timk", // questionable
			"B4GCSkORAJs",
			"D_4EDAqenHQ",
			"EICJp29EGOk",
			"Fxvd1timk",
			"GTqMZELYfQQ",
			"GZPTpLPEGmwHGWPC",
			"_HChnE9NDPY",
			"NwhjgIWHgGg",
			"production/tsbqksph4xswqjexfbec",
			"p/u/bguhrxupr23mw3nwxcrw",
			"nRSNapbJZnc",
			"zgfpbtolciznub5egzxk",
			"zufnu7aimadua9wrgwwo",
			"zznto1jzch9yjsbtbrul",
		*/
	}

	for _, str := range cases {
		if !IsGibberish(str) {
			t.Errorf("Mistakenly false: %s", str)
		}
	}
}

func TestVersionStrings(t *testing.T) {
	cases := []string{
		"1.0",
		"1.0.0",
		"v2.1.3",
		"2.1.73",
	}

	for _, str := range cases {
		if !IsVersionString(str) {
			t.Errorf("Mistakenly false: %s", str)
		}
	}
}
