#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdarg.h>
#include <errno.h>

/*
 * Print a string, escaping all required characters for clean HTML output.
 */
static void
print_htmlesc(const char *text)
{
	const uint8_t *buf;

	buf = (const uint8_t *)text;
	for (;;) {
		uint32_t x;

		x = *buf ++;
		if (x == 0) {
			return;
		}
		if (x >= 128) {
			int n;

			if (x < 0xC0) {
				x = 0xFFFD;
				n = 0;
			} else if (x < 0xE0) {
				x &= 0x1F;
				n = 1;
			} else if (x < 0xF0) {
				x &= 0x0F;
				n = 2;
			} else if (x < 0xF8) {
				x &= 0x07;
				n = 3;
			} else {
				x = 0xFFFD;
				n = 0;
			}
			while (n -- > 0) {
				uint32_t y;

				y = *buf ++;
				if (y >= 0x80 && y < 0xC0) {
					x = (x << 6) | (y & 0x3F);
				} else if (y == 0) {
					return;
				} else {
					x = 0xFFFD;
					break;
				}
			}
		}

		if (x >= 32 && x <= 126 && x != '<' && x != '>' && x != '&') {
			putchar((int)x);
		} else {
			printf("&x%X;", x);
		}
	}
}

/*
 * Send an error message translating a system error code.
 */
static void
send_system_error(const char *text, int err)
{
	printf("Status: 500 Failed\n");
	printf("Content-Type: text/html; charset=UTF-8\n");
	printf("\n");
	printf("<html>\n");
	printf("<body>\n");
	printf("<p>");
	print_htmlesc(text);
	printf(": errno = %d: ", err);
	print_htmlesc(strerror(err));
	printf("</p>\n");
	printf("</body>\n");
	printf("</html>\n");
}

int
main(void)
{
	FILE *f, *g;

	f = fopen("../show-marks.html", "r");
	if (f == NULL) {
		send_system_error("fopen() (show-marks.html)", errno);
		return 0;
	}
	g = fopen("../marks.txt", "r");
	if (g == NULL) {
		fclose(f);
		send_system_error("fopen() (marks.txt)", errno);
	}

	printf("Content-Type: text/html; charset=UTF-8\n");
	printf("\n");
	for (;;) {
		char line[512];
		size_t n;

		if (fgets(line, sizeof line, f) == NULL) {
			goto loop_exit;
		}
		n = strlen(line);
		if (n >= (sizeof line) - 1) {
			fputs(line, stdout);
			for (;;) {
				if (fgets(line, sizeof line, f) == NULL) {
					goto loop_exit;
				}
				fputs(line, stdout);
				if (strlen(line) < (sizeof line) - 1) {
					break;
				}
			}
			continue;
		}

		if (strstr(line, "MARKTABLE") == NULL) {
			fputs(line, stdout);
			continue;
		}

		printf("<table>\n");
		printf("<tr><th>Student Code</th><th>Mark</th><th>Comment</th></tr>\n");
		for (;;) {
			char *p, *comment, *e;
			long m;

			if (fgets(line, sizeof line, g) == NULL) {
				break;
			}
			p = strchr(line, ',');
			if (p == NULL) {
				continue;
			}
			*p ++ = 0;
			m = strtol(p, &e, 10);
			if (m < 0 || m > 100 || *p == 0 || *e >= 32) {
				comment = "INVALID";
			} else if (m < 30) {
				comment = "Poor";
			} else if (m < 60) {
				comment = "Average";
			} else if (m < 90) {
				comment = "Good";
			} else if (m < 100) {
				comment = "Very good";
			} else {
				comment = "PERFECT! Keep up the good work";
			}
			printf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n",
				line, p, comment);
		}
		printf("</table>\n");
	}
loop_exit:
	fclose(f);
	fclose(g);
	return 0;
}
