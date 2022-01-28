package oas

import "testing"

func TestNegative(t *testing.T) {
	cases := []string{
		"",
		"b", // can be valid hexadecimal
		"GetUniversalVariableUser",
		"callback",
		"runs",
		"tcfv2",
		"StartUpCheckout",
		"GetCart",
		"project-id",
		"data.json",
		"post.json",
		"test.png",
		"testdata-10kB.js",
		"g.js",
		"g.pixel",
		"opt-out",
		"profile-method-info",
		"GetAds",
		"fcgi-bin",
		".html",
		"agents.author.1.json",
		"publisha.1.json",
		"footer.include.html",
		"index.html",
		"Matt-cartoon-255x206px-small.png",
		"TheTelegraph_portal_white-320-small.png",
		"advert-management.adBlockerMessage.html",
		"Michael_Vaughan1.png",
		"big-danger-coronavirus-panic-greater-crisis",
		"some-uuid-maybe",
		"github-audit-exports",
		"dialog.overlay.infinity.json",
		"sync_a9",
		"1.0",
		"1.0.0",
		"v2.1.3",
		"image.sbix",
		"stable-4.0-version.json",
		"2.1.73",
		"zoom_in.cur",
		"pixel_details.html",
		"rtb-h",
		"fullHashes:find",
		"embeddable",
		"embeddable_blip",
		"abTestV2",
		"AddUserGroupLink",
		"web_widget",
		"VersionCheck.php",
		"{}",
	}

	for _, str := range cases {
		if IsGibberish(str) {
			t.Errorf("Mistakenly true: %s", str)
		}
	}
}

func TestPositive(t *testing.T) {
	cases := []string{
		"e21f7112-3d3b-4632-9da3-a4af2e0e9166",
		"952bea17-3776-11ea-9341-42010a84012a",
		"456795af-b48f-4a8d-9b37-3e932622c2f0",
		"0a0d0174-b338-4520-a1c3-24f7e3d5ec50.html",
		"6120c057c7a97b03f6986f1b",
		"610bc3fd5a77a7fa25033fb0",
		"610bd0315a77a7fa25034368",
		"610bd0315a77a7fa25034368zh",
		"710a462e",
		"1554507871",
		"19180481",
		"1024807212418223",
		"1553183382779",
		"qwerqwerasdfqwer@protonmai.com",
		"john.dow.1981@protonmail.com",
		"ci12NC01YzkyNTEzYzllMDRhLTAtYy5tb25pdG9yaW5nLmpzb24=", // long base64
		"11ca096cbc224a67360493d44a9903",
		"c738338322370b47a79251f7510dd", // prefixed hex
		"QgAAAC6zw0qH2DJtnXe8Z7rUJP0FgAFKkOhcHdFWzL1ZYggtwBgiB3LSoele9o3ZqFh7iCBhHbVLAnMuJ0HF8hEw7UKecE6wd-MBXgeRMdubGydhAMZSmuUjRpqplML40bmrb8VjJKNZswD1Cg",
		"QgAAAC6zw0qH2DJtnXe8Z7rUJP0rG4sjLa_KVLlww5WEDJ__30J15en-K_6Y68jb_rU93e2TFY6fb0MYiQ1UrLNMQufqODHZUl39Lo6cXAOVOThjAMZSmuVH7n85JOYSCgzpvowMAVueGG0Xxg",
		"203ef0f713abcebd8d62c35c0e3f12f87d71e5e4",
		"MDEyOk9yZ2FuaXphdGlvbjU3MzI0Nzk1",
		"730970532670-compute@developer.gserviceaccount.com",
		"arn-aws-ecs-eu-west-2-396248696294-cluster-london-01-ECSCluster-27iuIYva8nO4", // ?
		"AAAA028295945",
		"sp_ANQXRpqH_urn$3Auri$3Abase64$3A6698b0a3-97ad-52ce-8fc3-17d99e37a726",
		"n63nd45qsj",
		"n9z9QGNiz",
		"proxy.3d2100fd7107262ecb55ce6847f01fa5.html",
		"r-ext-5579e00a95c90",
		"r-ext-5579e8b12f11e",
		"r-v4-5c92513c9e04a",
		"r-v4-5c92513c9e04a-0-c.monitoring.json",
		"segments-1563566437171.639994",
		"t_52d94268-8810-4a7e-ba87-ffd657a6752f",
		"timeouts-1563566437171.639994",
		"a3226860758.html",
		"NC4WTmcy",

		// TODO
		// "fb6cjraf9cejut2a",
		// "Fxvd1timk", // questionable
		// "JEHJW4BKVFDRTMTUQLHKK5WVAU",
	}

	for _, str := range cases {
		if !IsGibberish(str) {
			t.Errorf("Mistakenly false: %s", str)
		}
	}
}
