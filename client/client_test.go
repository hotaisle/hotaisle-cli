package client

import "testing"

func TestBuildPath(t *testing.T) {
	tests := []struct {
		name     string
		template string
		params   map[string]string
		want     string
	}{
		{
			name:     "empty params map",
			template: "/teams/{team}/",
			params:   map[string]string{},
			want:     "/teams/{team}/",
		},
		{
			name:     "nil params map",
			template: "/teams/{team}/",
			params:   nil,
			want:     "/teams/{team}/",
		},
		{
			name:     "single parameter replacement",
			template: "/teams/{team}/",
			params:   map[string]string{"team": "my-team"},
			want:     "/teams/my-team/",
		},
		{
			name:     "multiple parameter replacements",
			template: "/teams/{team}/members/{email}/",
			params:   map[string]string{"team": "my-team", "email": "user@example.com"},
			want:     "/teams/my-team/members/user@example.com/",
		},
		{
			name:     "URL escaping special characters",
			template: "/teams/{team}/",
			params:   map[string]string{"team": "team with spaces"},
			want:     "/teams/team%20with%20spaces/",
		},
		{
			name:     "URL escaping slash",
			template: "/teams/{team}/",
			params:   map[string]string{"team": "team/name"},
			want:     "/teams/team%2Fname/",
		},
		{
			name:     "URL escaping special chars",
			template: "/user/{id}/",
			params:   map[string]string{"id": "user@example.com?sort=name"},
			want:     "/user/user@example.com%3Fsort=name/",
		},
		{
			name:     "no parameters in template",
			template: "/users/",
			params:   map[string]string{"team": "my-team"},
			want:     "/users/",
		},
		{
			name:     "parameter not in template",
			template: "/teams/{name}/",
			params:   map[string]string{"team": "my-team", "extra": "value"},
			want:     "/teams/{name}/",
		},
		{
			name:     "empty string parameter",
			template: "/teams/{team}/",
			params:   map[string]string{"team": ""},
			want:     "/teams//",
		},
		{
			name:     "unicode characters",
			template: "/teams/{team}/",
			params:   map[string]string{"team": "team-中文"},
			want:     "/teams/team-%E4%B8%AD%E6%96%87/",
		},
		{
			name:     "long template path with multiple params",
			template: "/teams/{team}/members/{email}/settings/{setting}/",
			params: map[string]string{
				"team":    "acme",
				"email":   "john@acme.com",
				"setting": "notifications",
			},
			want: "/teams/acme/members/john@acme.com/settings/notifications/",
		},
		{
			name:     "parameter with hyphens and underscores",
			template: "/virtual_machines/{vm_id}/",
			params:   map[string]string{"vm_id": "vm-123_abc"},
			want:     "/virtual_machines/vm-123_abc/",
		},
		{
			name:     "template without placeholders",
			template: "/api/v1/users",
			params:   map[string]string{"id": "123"},
			want:     "/api/v1/users",
		},
		{
			name:     "trailing slash preservation",
			template: "/teams/{team}/",
			params:   map[string]string{"team": "my-team"},
			want:     "/teams/my-team/",
		},
		{
			name:     "no trailing slash",
			template: "/teams/{team}",
			params:   map[string]string{"team": "my-team"},
			want:     "/teams/my-team",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildPath(tt.template, tt.params)
			if got != tt.want {
				t.Errorf("buildPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Benchmark to compare performance
func BenchmarkBuildPath(b *testing.B) {
	params := map[string]string{
		"team":  "my-team",
		"email": "user@example.com",
	}
	template := "/teams/{team}/members/{email}/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildPath(template, params)
	}
}

func BenchmarkBuildPathManyParams(b *testing.B) {
	params := map[string]string{
		"a": "value1",
		"b": "value2",
		"c": "value3",
		"d": "value4",
		"e": "value5",
	}
	template := "/path/{a}/to/{b}/resource/{c}/with/{d}/id/{e}/"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buildPath(template, params)
	}
}
