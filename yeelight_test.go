package yeelight

import (
	"testing"
)

func TestParseAddress(t *testing.T) {
	sample_input := `HTTP/1.1 200 OK\r\nCache-Control: max-age=3600\r\nDate: \r\nExt: \r\nLocation: yeelight://192.168.5.243:55443\r\n\r\nServer: POSIX UPnP/1.0 YGLC/1\r\nid: 0x000000001b31eeee\r\nmodel: colorb\r\nfw_ver: 10\r\nsupport: get_prop set_default set_power toggle set_bright set_scene cron_add cron_get cron_del start_cf stop_cf set_ct_abx adjust_ct set_name set_adjust adjust_bright adjust_color set_rgb set_hsv set_music udp_sess_new udp_sess_keep_alive udp_chroma_sess_new\r\npower: on\r\nbright: 100\r\ncolor_mode: 1\r\nct: 3200\r\nrgb: 16711680\r\nhue: 0\r\nsat: 100\r\nname: 
` + crlf

	expected := "192.168.5.243:55443"

	parsed, err := parseAddr(sample_input)

	if parsed != expected {
		t.Errorf("Address parsed as %v but expected %v. error: %s", parsed, expected, err)
	}
}
