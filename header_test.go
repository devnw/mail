package mail

import (
	"bufio"
	"context"
	"os"
	"strings"
	"testing"
)

var t1 = `Subject: Re: Test Subject 2
To: info@receiver.com
References: <2f6b7595-c01e-46e5-42bc-f263e1c4282d@receiver.com>
 <9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@domain.com>
Cc: Cc Man <ccman@gmail.com>
From: Sender Man <sender@domain.com>
Message-ID: <0e9a21b4-01dc-e5c1-dcd6-58ce5aa61f4f@receiver.com>
Date: Fri, 7 Apr 2017 12:59:55 +0200
User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:45.0)
 Gecko/20100101 Thunderbird/45.8.0
MIME-Version: 1.0
In-Reply-To: <9ff38d03-c4ab-89b7-9328-e99d5e24e3ba@receiver.eu>
Content-Type: multipart/alternative;
 boundary="------------C70C0458A558E585ACB75FB4"

This is a multi-part message in MIME format.
--------------C70C0458A558E585ACB75FB4
Content-Type: text/plain; charset=utf-8; format=flowed
Content-Transfer-Encoding: 8bit

First level
> Second level
>> Third level
>


--------------C70C0458A558E585ACB75FB4
Content-Type: multipart/related;
 boundary="------------5DB4A1356834BB602A5F88B2"


--------------5DB4A1356834BB602A5F88B2
Content-Type: text/html; charset=utf-8
Content-Transfer-Encoding: 8bit

<html>data<img src="part2.9599C449.04E5EC81@develhell.com"/></html>

--------------5DB4A1356834BB602A5F88B2
Content-Type: image/png
Content-Transfer-Encoding: base64
Content-ID: <part2.9599C449.04E5EC81@develhell.com>

iVBORw0KGgoAAAANSUhEUgAAAQEAAAAYCAIAAAB1IN9NAAAACXBIWXMAAAsTAAALEwEAmpwY
YKUKF+Os3baUndC0pDnwNAmLy1SUr2Gw0luxQuV/AwC6cEhVV5VRrwAAAABJRU5ErkJggg==
--------------5DB4A1356834BB602A5F88B2

--------------C70C0458A558E585ACB75FB4--
`

func Test_Header_Decode(t *testing.T) {
	e := &Email{}
	err := e.Decode(context.Background(), strings.NewReader(t1))
	if err != nil {
		t.Fatal(err)
	}

	for _, recv := range e.Received {
		t.Logf("%#v", recv)
	}
}

// TODO: add tests for '(from ' because apparently this is allowed

func Test_Header_Transform_Ignore(t *testing.T) {
	tests := []struct {
		input string
		err   error
	}{
		// https://metacpan.org/dist/Mail-SpamAssassin/source/lib/Mail/SpamAssassin/Message/Metadata/Received.pm#L343
		{
			"(qmail 27981 invoked by uid 225); 14 Mar 2003 07:24:34 -0000",
			ErrIgnoreTransport,
		},
		{
			"(qmail 84907 invoked from network); 13 Feb 2003 20:59:28 -0000",
			ErrIgnoreTransport,
		},
		{
			"(ofmipd 208.31.42.38); 17 Mar 2003 04:09:01 -0000",
			ErrIgnoreTransport,
		},
		{
			"by faerber.muc.de (OpenXP/32 v3.9.4 (Win32) alpha @ 2003-03-07-1751d); 07 Mar 2003 22:10:29 +0000",
			ErrIgnoreTransport,
		},
		{
			"by x.x.org (bulk_mailer v1.13); Wed, 26 Mar 2003 20:44:41 -0600",
			ErrIgnoreTransport,
		},
		{
			"by SPIDERMAN with Internet Mail Service (5.5.2653.19) id <19AF8VY2>; Tue, 25 Mar 2003 11:58:27 -0500",
			ErrIgnoreTransport,
		},
		{
			"by oak.ein.cz (Postfix, from userid 1002) id DABBD1BED3; Thu, 13 Feb 2003 14:02:21 +0100 (CET)",
			ErrIgnoreTransport,
		},
		{
			"OTM-MIX(otm-mix00) id k5N1aDtp040896; Fri, 23 Jun 2006 10:36:14 +0900 (JST)",
			ErrIgnoreTransport,
		},
		{
			"at Infodrom Oldenburg (/##/ Smail-3.2.0.102 1998-Aug-2 #2) from infodrom.org by finlandia.Infodrom.North.DE via smail from stdin id <m1FglM8-000okjC@finlandia.Infodrom.North.DE> for debian-security-announce@lists.debian.org; Thu, 18 May 2006 18:28:08 +0200 (CEST)",
			ErrIgnoreTransport,
		},
		{
			"with ECARTIS (v1.0.0; list bind-announce); Fri, 18 Aug 2006 07:19:58 +0000 (UTC)",
			ErrIgnoreTransport,
		},
		{
			"Message by Barricade wilhelm.eyp.ee with ESMTP id h1I7hGU06122 for <spamassassin-talk@lists.sourceforge.net>; Tue, 18 Feb 2003 09:43:16 +0200",
			ErrIgnoreTransport,
		},
		// https://metacpan.org/dist/Mail-SpamAssassin/source/lib/Mail/SpamAssassin/Message/Metadata/Received.pm#L359
		{
			"from www-data by wwwmail.documenta.de (Exim 4.50) with local for <example@vandinter.org> id 1GFbZc-0006QV-L8; Tue, 22 Aug 2006 21:06:04 +0200",
			ErrIgnoreTransport,
		},
		{
			"from server.yourhostingaccount.com with local  for example@vandinter.org  id 1GDtdl-0002GU-QE (8710); Thu, 17 Aug 2006 21:59:17 -0400",
			ErrIgnoreTransport,
		},
		// https://metacpan.org/dist/Mail-SpamAssassin/source/lib/Mail/SpamAssassin/Message/Metadata/Received.pm#L363
		{
			"from virtual-access.org by bolero.conactive.com ; Thu, 20 Feb 2003 23:32:58 +0100",
			ErrIgnoreTransport,
		},
		{
			"FROM ca-ex-bridge1.nai.com BY scwsout1.nai.com ; Fri Feb 07 10:18:12 2003 -0800",
			ErrIgnoreTransport,
		},
		// https://metacpan.org/dist/Mail-SpamAssassin/source/lib/Mail/SpamAssassin/Message/Metadata/Received.pm#L365
		{
			"from [86.122.158.69] by mta2.iomartmail.com; Thu, 2 Aug 2007 21:50:04 -0200",
			nil,
		},
		// https://metacpan.org/dist/Mail-SpamAssassin/source/lib/Mail/SpamAssassin/Message/Metadata/Received.pm#L374
		{
			"from av0001.technodiva.com (localhost [127.0.0.1])by  localhost.technodiva.com (Postfix) with ESMTP id 846CF2117for  <proftp-user@lists.sourceforge.net>; Mon,  7 Aug 2006 17:48:07 +0200 (MEST)",
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			trans := &Transport{}
			err := trans.Decode(context.Background(), test.input)
			if err != test.err {
				t.Fatalf("expected %s, got %v", test.err, err)
			}
		})
	}
}

func Test_Transport_Bulk(t *testing.T) {
	const data = "./testdata/received.txt"
	f, err := os.Open(data)
	if err != nil {
		t.Fatal(err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := strings.TrimPrefix(scanner.Text(), "Received: ")

		trans := &Transport{}
		err := trans.Decode(context.Background(), line)
		if err != nil && err != ErrIgnoreTransport {
			t.Fatalf("line: [%s] | error: %s", line, err)
		}
	}
}
