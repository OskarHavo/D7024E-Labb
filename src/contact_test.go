package main

import (
	"fmt"
	"net"
	"testing"
)

func TestNewContact(t *testing.T) {
	fmt.Println("This import is not unused >:(")
}

func TestContact_CalcDistance(t *testing.T) {
	type args struct {
		ID_1 *KademliaID
		ID_2 *KademliaID
	}
	tests := []struct {
		name string
		args args
		want *KademliaID
	}{
		{"equal distance",args{(*KademliaID)(make([]byte,ID_LEN)),(*KademliaID)(make([]byte,ID_LEN))},(*KademliaID)(make([]byte,ID_LEN))},
		{"1 bit distance",args{(*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1}),(*KademliaID)(make([]byte,ID_LEN))},(*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T){

			contactA := NewContact(tt.args.ID_1,"")

			contactA.CalcDistance(tt.args.ID_2)

			if !contactA.distance.Equals(tt.want) {
				t.Errorf("CalcDistance() = %v, want %v", contactA.distance.String(), tt.want.String())
			}
		})
	}
}

func TestContact_Less(t *testing.T) {
	type fields struct {
		ID       *KademliaID
		Address  string
		//distance *KademliaID
	}
	type args struct {
		otherContact Contact
		target *KademliaID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"not less",fields{(*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,255}),""},
			args{Contact{ID: (*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,1})},
				(*KademliaID)(make([]byte,ID_LEN))},
				false},
		{"not less",fields{NewKademliaID("0000000000000000000000000000000000FFFFFF"),""},
			args{Contact{ID:NewKademliaID("00000000000000000000000000000000000FFFFF")},
				(*KademliaID)(make([]byte,ID_LEN))},
			false},
		{"equal",fields{(*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,255}),""},
			args{Contact{ID: (*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,255})},
				(*KademliaID)(make([]byte,ID_LEN))},
			false},
		{"less",fields{(*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,16}),""},
			args{Contact{ID: (*KademliaID)([]byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,255})},
				(*KademliaID)(make([]byte,ID_LEN))},
			true},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contact := &Contact{
				ID:       tt.fields.ID,
				Address:  tt.fields.Address,
			}
			contact.CalcDistance(tt.args.target)
			tt.args.otherContact.CalcDistance(tt.args.target)
			if got := contact.Less(&tt.args.otherContact); got != tt.want {
				t.Errorf("Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContact_String(t *testing.T) {
	ip := net.ParseIP("0.0.0.0")
	type fields struct {
		ID       *KademliaID
		Address  string
		distance *KademliaID
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"",fields{NewKademliaIDFromIP(&ip),"0.0.0.0",NewKademliaIDFromData("h")},"contact(\"e562f69ec36e625116376f376d991e41613e9bf3\", \"0.0.0.0\")"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contact := &Contact{
				ID:       tt.fields.ID,
				Address:  tt.fields.Address,
				distance: tt.fields.distance,
			}
			if got := contact.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContactCandidates_Append(t *testing.T) {
	type fields struct {
		contacts []Contact
	}
	type args struct {
		contacts []Contact
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{ "+1 contact",fields{[]Contact{}},args{contacts: []Contact{{}}}},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &ContactCandidates{
				contacts: tt.fields.contacts,
			}
			length := candidates.Len()
			candidates.Append(tt.args.contacts)
			if length == candidates.Len() {
				length += len(tt.args.contacts)
				t.Errorf("Append() = %v, want %v", candidates.Len(),length)
			}
		})
	}
}

func TestContactCandidates_AppendContact(t *testing.T) {
	type args struct {
		contact Contact
	}
	tests := []struct {
		name   string
		args args
	}{
		{"+1 contact", args{contact: Contact{}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			candidates := &ContactCandidates{
				contacts: []Contact{},
			}
			len := candidates.Len()

			candidates.AppendContact(tt.args.contact)

			if len == candidates.Len() {
				t.Errorf("AppendContact() = %v, want %v", candidates.Len(), len+1)
			}

		})
	}
}

func TestContactCandidates_GetContacts(t *testing.T) {
	type fields struct {
		contacts []Contact
	}
	type args struct {
		count int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{"Get 1 contact",fields{[]Contact{{}}},args{1},1},
		{"Get 0 contacts",fields{[]Contact{}},args{1},0},
		{"Get <2 contacts",fields{[]Contact{{}}},args{2},1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &ContactCandidates{
				contacts: tt.fields.contacts,
			}
			if temp_len := len(candidates.GetContacts(tt.args.count)); temp_len != tt.want {
				t.Errorf("GetContacts() = %v, want %v", temp_len, tt.want)
			}
		})
	}
}

func TestContactCandidates_Len(t *testing.T) {
	type fields struct {
		contacts []Contact
	}
	tests := []struct {
		name   string
		fields fields
		want   int
	}{
		{"Empty",fields{[]Contact{}},0},
		{"1 contact",fields{[]Contact{{}}},1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &ContactCandidates{
				contacts: tt.fields.contacts,
			}
			if got := candidates.Len(); got != tt.want {
				t.Errorf("Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContactCandidates_Swap(t *testing.T) {
	type fields struct {
		contacts []Contact
	}
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Contact swap among 2",fields{[]Contact{{ID: NewKademliaIDFromData("String 1")},{ID: NewKademliaIDFromData("String 2")}}},args{0,1}},
		{"Contact swap among 4",fields{[]Contact{
			{ID: NewKademliaIDFromData("String 1")},
			{ID: NewKademliaIDFromData("String 2")},
			{ID: NewKademliaIDFromData("String 3")},
			{ID: NewKademliaIDFromData("String 4")}}},args{1,2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &ContactCandidates{
				contacts: tt.fields.contacts,
			}

			ID_1 := candidates.contacts[tt.args.i].ID
			ID_2 := candidates.contacts[tt.args.j].ID
			candidates.Swap(tt.args.i,tt.args.j)
			new_ID_1 := candidates.contacts[tt.args.i].ID
			new_ID_2 := candidates.contacts[tt.args.j].ID
			if !new_ID_1.Equals(ID_2) || !new_ID_2.Equals(ID_1) {
				t.Errorf("Len() = %v, want %v and Len() = %v, want %v", new_ID_1.String(), ID_2.String(),new_ID_2.String(),ID_1.String())
			}
		})
	}
}

func TestContactCandidates_Contains(t *testing.T) {
	type fields struct {
		contacts []Contact
	}
	type args struct {
		contact Contact
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{"Does not contain",fields{[]Contact{}},args{contact: Contact{ID: NewKademliaIDFromData("my ID")}},false},
		{"Does not contain",fields{[]Contact{}},args{contact: Contact{ID: NewKademliaIDFromData("my ID")}},false},
		{"Contains",fields{[]Contact{{ID: NewKademliaIDFromData("my ID")}}},args{contact: Contact{ID: NewKademliaIDFromData("my ID")}},true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &ContactCandidates{
				contacts: tt.fields.contacts,
			}
			if got := candidates.Contains(&tt.args.contact); got != tt.want {
				t.Errorf("Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContactCandidates_Remove(t *testing.T) {
	type fields struct {
		contacts []Contact
	}
	type args struct {
		contact Contact
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wants int
	}{
		{"Remove",fields{[]Contact{{ID: NewKademliaIDFromData("my ID")}}},args{contact: Contact{ID: NewKademliaIDFromData("my ID")}},0},
		{"Remove",fields{[]Contact{{ID: NewKademliaIDFromData("my ID")}}},args{contact: Contact{ID: NewKademliaIDFromData("my other ID")}},1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidates := &ContactCandidates{
				contacts: tt.fields.contacts,
			}
			candidates.Remove(&tt.args.contact)
			length := candidates.Len()

			if length != tt.wants {
				t.Errorf("Remove() = %v, want %v", length, tt.wants)
			}

		})
	}
}

