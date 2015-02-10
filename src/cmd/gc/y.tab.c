/* A Bison parser, made by GNU Bison 2.3.  */

/* Skeleton implementation for Bison's Yacc-like parsers in C

   Copyright (C) 1984, 1989, 1990, 2000, 2001, 2002, 2003, 2004, 2005, 2006
   Free Software Foundation, Inc.

   This program is free software; you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation; either version 2, or (at your option)
   any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program; if not, write to the Free Software
   Foundation, Inc., 51 Franklin Street, Fifth Floor,
   Boston, MA 02110-1301, USA.  */

/* As a special exception, you may create a larger work that contains
   part or all of the Bison parser skeleton and distribute that work
   under terms of your choice, so long as that work isn't itself a
   parser generator using the skeleton or a modified version thereof
   as a parser skeleton.  Alternatively, if you modify or redistribute
   the parser skeleton itself, you may (at your option) remove this
   special exception, which will cause the skeleton and the resulting
   Bison output files to be licensed under the GNU General Public
   License without this special exception.

   This special exception was added by the Free Software Foundation in
   version 2.2 of Bison.  */

/* C LALR(1) parser skeleton written by Richard Stallman, by
   simplifying the original so-called "semantic" parser.  */

/* All symbols defined below should begin with yy or YY, to avoid
   infringing on user name space.  This should be done even for local
   variables, as they might otherwise be expanded by user macros.
   There are some unavoidable exceptions within include files to
   define necessary library symbols; they are noted "INFRINGES ON
   USER NAME SPACE" below.  */

/* Identify Bison output.  */
#define YYBISON 1

/* Bison version.  */
#define YYBISON_VERSION "2.3"

/* Skeleton name.  */
#define YYSKELETON_NAME "yacc.c"

/* Pure parsers.  */
#define YYPURE 0

/* Using locations.  */
#define YYLSP_NEEDED 0



/* Tokens.  */
#ifndef YYTOKENTYPE
# define YYTOKENTYPE
   /* Put the tokens into the symbol table, so that GDB and other debuggers
      know about them.  */
   enum yytokentype {
     LLITERAL = 258,
     LASOP = 259,
     LCOLAS = 260,
     LBREAK = 261,
     LCASE = 262,
     LCHAN = 263,
     LCONST = 264,
     LCONTINUE = 265,
     LDDD = 266,
     LDEFAULT = 267,
     LDEFER = 268,
     LELSE = 269,
     LFALL = 270,
     LFOR = 271,
     LFUNC = 272,
     LGO = 273,
     LGOTO = 274,
     LIF = 275,
     LIMPORT = 276,
     LINTERFACE = 277,
     LMAP = 278,
     LNAME = 279,
     LPACKAGE = 280,
     LRANGE = 281,
     LRETURN = 282,
     LSELECT = 283,
     LSTRUCT = 284,
     LSWITCH = 285,
     LTYPE = 286,
     LVAR = 287,
     LANDAND = 288,
     LANDNOT = 289,
     LBODY = 290,
     LCOMM = 291,
     LDEC = 292,
     LEQ = 293,
     LGE = 294,
     LGT = 295,
     LIGNORE = 296,
     LINC = 297,
     LLE = 298,
     LLSH = 299,
     LLT = 300,
     LNE = 301,
     LOROR = 302,
     LRSH = 303,
     NotPackage = 304,
     NotParen = 305,
     PreferToRightParen = 306
   };
#endif
/* Tokens.  */
#define LLITERAL 258
#define LASOP 259
#define LCOLAS 260
#define LBREAK 261
#define LCASE 262
#define LCHAN 263
#define LCONST 264
#define LCONTINUE 265
#define LDDD 266
#define LDEFAULT 267
#define LDEFER 268
#define LELSE 269
#define LFALL 270
#define LFOR 271
#define LFUNC 272
#define LGO 273
#define LGOTO 274
#define LIF 275
#define LIMPORT 276
#define LINTERFACE 277
#define LMAP 278
#define LNAME 279
#define LPACKAGE 280
#define LRANGE 281
#define LRETURN 282
#define LSELECT 283
#define LSTRUCT 284
#define LSWITCH 285
#define LTYPE 286
#define LVAR 287
#define LANDAND 288
#define LANDNOT 289
#define LBODY 290
#define LCOMM 291
#define LDEC 292
#define LEQ 293
#define LGE 294
#define LGT 295
#define LIGNORE 296
#define LINC 297
#define LLE 298
#define LLSH 299
#define LLT 300
#define LNE 301
#define LOROR 302
#define LRSH 303
#define NotPackage 304
#define NotParen 305
#define PreferToRightParen 306




/* Copy the first part of user declarations.  */
#line 20 "go.y"

#include <u.h>
#include <stdio.h>	/* if we don't, bison will, and go.h re-#defines getc */
#include <libc.h>
#include "go.h"

static void fixlbrace(int);


/* Enabling traces.  */
#ifndef YYDEBUG
# define YYDEBUG 0
#endif

/* Enabling verbose error messages.  */
#ifdef YYERROR_VERBOSE
# undef YYERROR_VERBOSE
# define YYERROR_VERBOSE 1
#else
# define YYERROR_VERBOSE 1
#endif

/* Enabling the token table.  */
#ifndef YYTOKEN_TABLE
# define YYTOKEN_TABLE 0
#endif

#if ! defined YYSTYPE && ! defined YYSTYPE_IS_DECLARED
typedef union YYSTYPE
#line 28 "go.y"
{
	Node*		node;
	NodeList*		list;
	Type*		type;
	Sym*		sym;
	struct	Val	val;
	int		i;
}
/* Line 193 of yacc.c.  */
#line 216 "y.tab.c"
	YYSTYPE;
# define yystype YYSTYPE /* obsolescent; will be withdrawn */
# define YYSTYPE_IS_DECLARED 1
# define YYSTYPE_IS_TRIVIAL 1
#endif



/* Copy the second part of user declarations.  */


/* Line 216 of yacc.c.  */
#line 229 "y.tab.c"

#ifdef short
# undef short
#endif

#ifdef YYTYPE_UINT8
typedef YYTYPE_UINT8 yytype_uint8;
#else
typedef unsigned char yytype_uint8;
#endif

#ifdef YYTYPE_INT8
typedef YYTYPE_INT8 yytype_int8;
#elif (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
typedef signed char yytype_int8;
#else
typedef short int yytype_int8;
#endif

#ifdef YYTYPE_UINT16
typedef YYTYPE_UINT16 yytype_uint16;
#else
typedef unsigned short int yytype_uint16;
#endif

#ifdef YYTYPE_INT16
typedef YYTYPE_INT16 yytype_int16;
#else
typedef short int yytype_int16;
#endif

#ifndef YYSIZE_T
# ifdef __SIZE_TYPE__
#  define YYSIZE_T __SIZE_TYPE__
# elif defined size_t
#  define YYSIZE_T size_t
# elif ! defined YYSIZE_T && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
#  include <stddef.h> /* INFRINGES ON USER NAME SPACE */
#  define YYSIZE_T size_t
# else
#  define YYSIZE_T unsigned int
# endif
#endif

#define YYSIZE_MAXIMUM ((YYSIZE_T) -1)

#ifndef YY_
# if defined YYENABLE_NLS && YYENABLE_NLS
#  if ENABLE_NLS
#   include <libintl.h> /* INFRINGES ON USER NAME SPACE */
#   define YY_(msgid) dgettext ("bison-runtime", msgid)
#  endif
# endif
# ifndef YY_
#  define YY_(msgid) msgid
# endif
#endif

/* Suppress unused-variable warnings by "using" E.  */
#if ! defined lint || defined __GNUC__
# define YYUSE(e) ((void) (e))
#else
# define YYUSE(e) /* empty */
#endif

/* Identity function, used to suppress warnings about constant conditions.  */
#ifndef lint
# define YYID(n) (n)
#else
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static int
YYID (int i)
#else
static int
YYID (i)
    int i;
#endif
{
  return i;
}
#endif

#if ! defined yyoverflow || YYERROR_VERBOSE

/* The parser invokes alloca or malloc; define the necessary symbols.  */

# ifdef YYSTACK_USE_ALLOCA
#  if YYSTACK_USE_ALLOCA
#   ifdef __GNUC__
#    define YYSTACK_ALLOC __builtin_alloca
#   elif defined __BUILTIN_VA_ARG_INCR
#    include <alloca.h> /* INFRINGES ON USER NAME SPACE */
#   elif defined _AIX
#    define YYSTACK_ALLOC __alloca
#   elif defined _MSC_VER
#    include <malloc.h> /* INFRINGES ON USER NAME SPACE */
#    define alloca _alloca
#   else
#    define YYSTACK_ALLOC alloca
#    if ! defined _ALLOCA_H && ! defined _STDLIB_H && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
#     include <stdlib.h> /* INFRINGES ON USER NAME SPACE */
#     ifndef _STDLIB_H
#      define _STDLIB_H 1
#     endif
#    endif
#   endif
#  endif
# endif

# ifdef YYSTACK_ALLOC
   /* Pacify GCC's `empty if-body' warning.  */
#  define YYSTACK_FREE(Ptr) do { /* empty */; } while (YYID (0))
#  ifndef YYSTACK_ALLOC_MAXIMUM
    /* The OS might guarantee only one guard page at the bottom of the stack,
       and a page size can be as small as 4096 bytes.  So we cannot safely
       invoke alloca (N) if N exceeds 4096.  Use a slightly smaller number
       to allow for a few compiler-allocated temporary stack slots.  */
#   define YYSTACK_ALLOC_MAXIMUM 4032 /* reasonable circa 2006 */
#  endif
# else
#  define YYSTACK_ALLOC YYMALLOC
#  define YYSTACK_FREE YYFREE
#  ifndef YYSTACK_ALLOC_MAXIMUM
#   define YYSTACK_ALLOC_MAXIMUM YYSIZE_MAXIMUM
#  endif
#  if (defined __cplusplus && ! defined _STDLIB_H \
       && ! ((defined YYMALLOC || defined malloc) \
	     && (defined YYFREE || defined free)))
#   include <stdlib.h> /* INFRINGES ON USER NAME SPACE */
#   ifndef _STDLIB_H
#    define _STDLIB_H 1
#   endif
#  endif
#  ifndef YYMALLOC
#   define YYMALLOC malloc
#   if ! defined malloc && ! defined _STDLIB_H && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
void *malloc (YYSIZE_T); /* INFRINGES ON USER NAME SPACE */
#   endif
#  endif
#  ifndef YYFREE
#   define YYFREE free
#   if ! defined free && ! defined _STDLIB_H && (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
void free (void *); /* INFRINGES ON USER NAME SPACE */
#   endif
#  endif
# endif
#endif /* ! defined yyoverflow || YYERROR_VERBOSE */


#if (! defined yyoverflow \
     && (! defined __cplusplus \
	 || (defined YYSTYPE_IS_TRIVIAL && YYSTYPE_IS_TRIVIAL)))

/* A type that is properly aligned for any stack member.  */
union yyalloc
{
  yytype_int16 yyss;
  YYSTYPE yyvs;
  };

/* The size of the maximum gap between one aligned stack and the next.  */
# define YYSTACK_GAP_MAXIMUM (sizeof (union yyalloc) - 1)

/* The size of an array large to enough to hold all stacks, each with
   N elements.  */
# define YYSTACK_BYTES(N) \
     ((N) * (sizeof (yytype_int16) + sizeof (YYSTYPE)) \
      + YYSTACK_GAP_MAXIMUM)

/* Copy COUNT objects from FROM to TO.  The source and destination do
   not overlap.  */
# ifndef YYCOPY
#  if defined __GNUC__ && 1 < __GNUC__
#   define YYCOPY(To, From, Count) \
      __builtin_memcpy (To, From, (Count) * sizeof (*(From)))
#  else
#   define YYCOPY(To, From, Count)		\
      do					\
	{					\
	  YYSIZE_T yyi;				\
	  for (yyi = 0; yyi < (Count); yyi++)	\
	    (To)[yyi] = (From)[yyi];		\
	}					\
      while (YYID (0))
#  endif
# endif

/* Relocate STACK from its old location to the new one.  The
   local variables YYSIZE and YYSTACKSIZE give the old and new number of
   elements in the stack, and YYPTR gives the new location of the
   stack.  Advance YYPTR to a properly aligned location for the next
   stack.  */
# define YYSTACK_RELOCATE(Stack)					\
    do									\
      {									\
	YYSIZE_T yynewbytes;						\
	YYCOPY (&yyptr->Stack, Stack, yysize);				\
	Stack = &yyptr->Stack;						\
	yynewbytes = yystacksize * sizeof (*Stack) + YYSTACK_GAP_MAXIMUM; \
	yyptr += yynewbytes / sizeof (*yyptr);				\
      }									\
    while (YYID (0))

#endif

/* YYFINAL -- State number of the termination state.  */
#define YYFINAL  5
/* YYLAST -- Last index in YYTABLE.  */
#define YYLAST   2383

/* YYNTOKENS -- Number of terminals.  */
#define YYNTOKENS  76
/* YYNNTS -- Number of nonterminals.  */
#define YYNNTS  191
/* YYNRULES -- Number of rules.  */
#define YYNRULES  401
/* YYNRULES -- Number of states.  */
#define YYNSTATES  766

/* YYTRANSLATE(YYLEX) -- Bison symbol number corresponding to YYLEX.  */
#define YYUNDEFTOK  2
#define YYMAXUTOK   306

#define YYTRANSLATE(YYX)						\
  ((unsigned int) (YYX) <= YYMAXUTOK ? yytranslate[YYX] : YYUNDEFTOK)

/* YYTRANSLATE[YYLEX] -- Bison symbol number corresponding to YYLEX.  */
static const yytype_uint8 yytranslate[] =
{
       0,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    69,     2,     2,    64,    55,    56,     2,
      59,    60,    53,    49,    75,    50,    63,    54,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,    66,    62,
       2,    65,     2,    73,    74,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,    71,     2,    72,    52,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,    67,    51,    68,    70,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     2,     2,     2,     2,
       2,     2,     2,     2,     2,     2,     1,     2,     3,     4,
       5,     6,     7,     8,     9,    10,    11,    12,    13,    14,
      15,    16,    17,    18,    19,    20,    21,    22,    23,    24,
      25,    26,    27,    28,    29,    30,    31,    32,    33,    34,
      35,    36,    37,    38,    39,    40,    41,    42,    43,    44,
      45,    46,    47,    48,    57,    58,    61
};

#if YYDEBUG
/* YYPRHS[YYN] -- Index of the first RHS symbol of rule number YYN in
   YYRHS.  */
static const yytype_uint16 yyprhs[] =
{
       0,     0,     3,     8,     9,    13,    39,    40,    44,    45,
      49,    50,    54,    55,    59,    60,    64,    65,    69,    70,
      74,    75,    79,    80,    84,    85,    89,    90,    94,    95,
      99,   100,   104,   105,   109,   110,   114,   115,   119,   120,
     124,   125,   129,   130,   134,   135,   139,   140,   144,   145,
     149,   150,   154,   155,   159,   160,   164,   165,   169,   172,
     178,   182,   186,   189,   191,   195,   197,   200,   203,   208,
     209,   211,   212,   217,   218,   220,   222,   224,   226,   229,
     235,   239,   242,   248,   256,   260,   263,   269,   273,   275,
     278,   283,   287,   292,   296,   298,   301,   303,   305,   308,
     310,   314,   318,   322,   325,   328,   332,   338,   344,   347,
     348,   353,   354,   358,   359,   362,   363,   368,   373,   378,
     381,   387,   389,   391,   394,   395,   399,   401,   405,   406,
     407,   408,   417,   418,   424,   425,   428,   429,   432,   433,
     434,   442,   443,   449,   451,   455,   459,   463,   467,   471,
     475,   479,   483,   487,   491,   495,   499,   503,   507,   511,
     515,   519,   523,   527,   531,   533,   536,   539,   542,   545,
     548,   551,   554,   557,   561,   567,   574,   576,   578,   582,
     588,   594,   599,   606,   615,   617,   623,   629,   635,   643,
     645,   646,   650,   652,   657,   659,   664,   666,   670,   672,
     674,   676,   678,   680,   682,   684,   685,   687,   689,   691,
     693,   698,   703,   705,   707,   709,   712,   714,   716,   718,
     720,   722,   726,   728,   730,   732,   735,   737,   739,   741,
     743,   747,   749,   751,   753,   755,   757,   759,   761,   763,
     765,   769,   774,   779,   782,   786,   792,   794,   796,   799,
     803,   809,   813,   819,   823,   827,   833,   842,   848,   857,
     863,   864,   868,   869,   871,   875,   877,   882,   885,   886,
     890,   892,   896,   898,   902,   904,   908,   910,   914,   916,
     920,   924,   927,   932,   936,   942,   948,   950,   954,   956,
     959,   961,   965,   970,   972,   975,   978,   980,   982,   986,
     987,   990,   991,   993,   995,   997,   999,  1001,  1003,  1005,
    1007,  1009,  1010,  1015,  1017,  1020,  1023,  1026,  1029,  1032,
    1035,  1037,  1041,  1043,  1047,  1049,  1053,  1055,  1059,  1061,
    1065,  1067,  1069,  1073,  1077,  1078,  1081,  1082,  1084,  1085,
    1087,  1088,  1090,  1091,  1093,  1094,  1096,  1097,  1099,  1100,
    1102,  1103,  1105,  1106,  1108,  1113,  1118,  1124,  1131,  1136,
    1141,  1143,  1145,  1147,  1149,  1151,  1153,  1155,  1157,  1159,
    1163,  1168,  1174,  1179,  1184,  1187,  1190,  1195,  1199,  1203,
    1209,  1213,  1218,  1222,  1228,  1230,  1231,  1233,  1237,  1239,
    1241,  1244,  1246,  1248,  1254,  1255,  1258,  1260,  1264,  1266,
    1270,  1272
};

/* YYRHS -- A `-1'-separated list of the rules' RHS.  */
static const yytype_int16 yyrhs[] =
{
      77,     0,    -1,    79,    78,   130,   215,    -1,    -1,    25,
     190,    62,    -1,    80,    82,    84,    86,    88,    90,    92,
      94,    96,    98,   100,   102,   104,   106,   108,   110,   112,
     114,   116,   118,   120,   122,   124,   126,   128,    -1,    -1,
      81,   135,   137,    -1,    -1,    83,   135,   137,    -1,    -1,
      85,   135,   137,    -1,    -1,    87,   135,   137,    -1,    -1,
      89,   135,   137,    -1,    -1,    91,   135,   137,    -1,    -1,
      93,   135,   137,    -1,    -1,    95,   135,   137,    -1,    -1,
      97,   135,   137,    -1,    -1,    99,   135,   137,    -1,    -1,
     101,   135,   137,    -1,    -1,   103,   135,   137,    -1,    -1,
     105,   135,   137,    -1,    -1,   107,   135,   137,    -1,    -1,
     109,   135,   137,    -1,    -1,   111,   135,   137,    -1,    -1,
     113,   135,   137,    -1,    -1,   115,   135,   137,    -1,    -1,
     117,   135,   137,    -1,    -1,   119,   135,   137,    -1,    -1,
     121,   135,   137,    -1,    -1,   123,   135,   137,    -1,    -1,
     125,   135,   137,    -1,    -1,   127,   135,   137,    -1,    -1,
     129,   135,   137,    -1,    -1,   130,   131,    62,    -1,    21,
     132,    -1,    21,    59,   133,   239,    60,    -1,    21,    59,
      60,    -1,   134,   135,   137,    -1,   134,   137,    -1,   132,
      -1,   133,    62,   132,    -1,     3,    -1,   190,     3,    -1,
      63,     3,    -1,    25,    24,   136,    62,    -1,    -1,    24,
      -1,    -1,   138,   263,    64,    64,    -1,    -1,   140,    -1,
     207,    -1,   230,    -1,     1,    -1,    32,   142,    -1,    32,
      59,   216,   239,    60,    -1,    32,    59,    60,    -1,   141,
     143,    -1,   141,    59,   143,   239,    60,    -1,   141,    59,
     143,    62,   217,   239,    60,    -1,   141,    59,    60,    -1,
      31,   146,    -1,    31,    59,   218,   239,    60,    -1,    31,
      59,    60,    -1,     9,    -1,   234,   195,    -1,   234,   195,
      65,   235,    -1,   234,    65,   235,    -1,   234,   195,    65,
     235,    -1,   234,    65,   235,    -1,   143,    -1,   234,   195,
      -1,   234,    -1,   190,    -1,   145,   195,    -1,   175,    -1,
     175,     4,   175,    -1,   235,    65,   235,    -1,   235,     5,
     235,    -1,   175,    42,    -1,   175,    37,    -1,     7,   236,
      66,    -1,     7,   236,    65,   175,    66,    -1,     7,   236,
       5,   175,    66,    -1,    12,    66,    -1,    -1,    67,   150,
     232,    68,    -1,    -1,   148,   152,   232,    -1,    -1,   153,
     151,    -1,    -1,    35,   155,   232,    68,    -1,   235,    65,
      26,   175,    -1,   235,     5,    26,   175,    -1,    26,   175,
      -1,   243,    62,   243,    62,   243,    -1,   243,    -1,   156,
      -1,   157,   154,    -1,    -1,    16,   160,   158,    -1,   243,
      -1,   243,    62,   243,    -1,    -1,    -1,    -1,    20,   163,
     161,   164,   154,   165,   168,   169,    -1,    -1,    14,    20,
     167,   161,   154,    -1,    -1,   168,   166,    -1,    -1,    14,
     149,    -1,    -1,    -1,    30,   171,   161,   172,    35,   153,
      68,    -1,    -1,    28,   174,    35,   153,    68,    -1,   176,
      -1,   175,    47,   175,    -1,   175,    33,   175,    -1,   175,
      38,   175,    -1,   175,    46,   175,    -1,   175,    45,   175,
      -1,   175,    43,   175,    -1,   175,    39,   175,    -1,   175,
      40,   175,    -1,   175,    49,   175,    -1,   175,    50,   175,
      -1,   175,    51,   175,    -1,   175,    52,   175,    -1,   175,
      53,   175,    -1,   175,    54,   175,    -1,   175,    55,   175,
      -1,   175,    56,   175,    -1,   175,    34,   175,    -1,   175,
      44,   175,    -1,   175,    48,   175,    -1,   175,    36,   175,
      -1,   183,    -1,    53,   176,    -1,    56,   176,    -1,    49,
     176,    -1,    50,   176,    -1,    69,   176,    -1,    70,   176,
      -1,    52,   176,    -1,    36,   176,    -1,   183,    59,    60,
      -1,   183,    59,   236,   240,    60,    -1,   183,    59,   236,
      11,   240,    60,    -1,     3,    -1,   192,    -1,   183,    63,
     190,    -1,   183,    63,    59,   184,    60,    -1,   183,    63,
      59,    31,    60,    -1,   183,    71,   175,    72,    -1,   183,
      71,   241,    66,   241,    72,    -1,   183,    71,   241,    66,
     241,    66,   241,    72,    -1,   177,    -1,   198,    59,   175,
     240,    60,    -1,   199,   186,   179,   238,    68,    -1,   178,
      67,   179,   238,    68,    -1,    59,   184,    60,    67,   179,
     238,    68,    -1,   214,    -1,    -1,   175,    66,   182,    -1,
     175,    -1,    67,   179,   238,    68,    -1,   175,    -1,    67,
     179,   238,    68,    -1,   178,    -1,    59,   184,    60,    -1,
     175,    -1,   196,    -1,   195,    -1,    35,    -1,    67,    -1,
     190,    -1,   190,    -1,    -1,   187,    -1,    24,    -1,   191,
      -1,    73,    -1,    74,     3,    63,    24,    -1,    74,     3,
      63,    73,    -1,   190,    -1,   187,    -1,    11,    -1,    11,
     195,    -1,   204,    -1,   210,    -1,   202,    -1,   203,    -1,
     201,    -1,    59,   195,    60,    -1,   204,    -1,   210,    -1,
     202,    -1,    53,   196,    -1,   210,    -1,   202,    -1,   203,
      -1,   201,    -1,    59,   195,    60,    -1,   210,    -1,   202,
      -1,   202,    -1,   204,    -1,   210,    -1,   202,    -1,   203,
      -1,   201,    -1,   192,    -1,   192,    63,   190,    -1,    71,
     241,    72,   195,    -1,    71,    11,    72,   195,    -1,     8,
     197,    -1,     8,    36,   195,    -1,    23,    71,   195,    72,
     195,    -1,   205,    -1,   206,    -1,    53,   195,    -1,    36,
       8,   195,    -1,    29,   186,   219,   239,    68,    -1,    29,
     186,    68,    -1,    22,   186,   220,   239,    68,    -1,    22,
     186,    68,    -1,    17,   208,   211,    -1,   190,    59,   228,
      60,   212,    -1,    59,   228,    60,   190,    59,   228,    60,
     212,    -1,   249,    59,   244,    60,   259,    -1,    59,   264,
      60,   190,    59,   244,    60,   259,    -1,    17,    59,   228,
      60,   212,    -1,    -1,    67,   232,    68,    -1,    -1,   200,
      -1,    59,   228,    60,    -1,   210,    -1,   213,   186,   232,
      68,    -1,   213,     1,    -1,    -1,   215,   139,    62,    -1,
     142,    -1,   216,    62,   142,    -1,   144,    -1,   217,    62,
     144,    -1,   146,    -1,   218,    62,   146,    -1,   221,    -1,
     219,    62,   221,    -1,   224,    -1,   220,    62,   224,    -1,
     233,   195,   247,    -1,   223,   247,    -1,    59,   223,    60,
     247,    -1,    53,   223,   247,    -1,    59,    53,   223,    60,
     247,    -1,    53,    59,   223,    60,   247,    -1,    24,    -1,
      24,    63,   190,    -1,   222,    -1,   187,   225,    -1,   222,
      -1,    59,   222,    60,    -1,    59,   228,    60,   212,    -1,
     185,    -1,   190,   185,    -1,   190,   194,    -1,   194,    -1,
     226,    -1,   227,    75,   226,    -1,    -1,   227,   240,    -1,
      -1,   149,    -1,   140,    -1,   230,    -1,     1,    -1,   147,
      -1,   159,    -1,   170,    -1,   173,    -1,   162,    -1,    -1,
     193,    66,   231,   229,    -1,    15,    -1,     6,   189,    -1,
      10,   189,    -1,    18,   177,    -1,    13,   177,    -1,    19,
     187,    -1,    27,   242,    -1,   229,    -1,   232,    62,   229,
      -1,   187,    -1,   233,    75,   187,    -1,   188,    -1,   234,
      75,   188,    -1,   175,    -1,   235,    75,   175,    -1,   184,
      -1,   236,    75,   184,    -1,   180,    -1,   181,    -1,   237,
      75,   180,    -1,   237,    75,   181,    -1,    -1,   237,   240,
      -1,    -1,    62,    -1,    -1,    75,    -1,    -1,   175,    -1,
      -1,   235,    -1,    -1,   147,    -1,    -1,   264,    -1,    -1,
     265,    -1,    -1,   266,    -1,    -1,     3,    -1,    21,    24,
       3,    62,    -1,    32,   249,   251,    62,    -1,     9,   249,
      65,   262,    62,    -1,     9,   249,   251,    65,   262,    62,
      -1,    31,   250,   251,    62,    -1,    17,   209,   211,    62,
      -1,   191,    -1,   249,    -1,   253,    -1,   254,    -1,   255,
      -1,   253,    -1,   255,    -1,   191,    -1,    24,    -1,    71,
      72,   251,    -1,    71,     3,    72,   251,    -1,    23,    71,
     251,    72,   251,    -1,    29,    67,   245,    68,    -1,    22,
      67,   246,    68,    -1,    53,   251,    -1,     8,   252,    -1,
       8,    59,   254,    60,    -1,     8,    36,   251,    -1,    36,
       8,   251,    -1,    17,    59,   244,    60,   259,    -1,   190,
     251,   247,    -1,   190,    11,   251,   247,    -1,   190,   251,
     247,    -1,   190,    59,   244,    60,   259,    -1,   251,    -1,
      -1,   260,    -1,    59,   244,    60,    -1,   251,    -1,     3,
      -1,    50,     3,    -1,   190,    -1,   261,    -1,    59,   261,
      49,   261,    60,    -1,    -1,   263,   248,    -1,   256,    -1,
     264,    75,   256,    -1,   257,    -1,   265,    62,   257,    -1,
     258,    -1,   266,    62,   258,    -1
};

/* YYRLINE[YYN] -- source line where rule number YYN was defined.  */
static const yytype_uint16 yyrline[] =
{
       0,   124,   124,   133,   139,   150,   179,   179,   196,   196,
     213,   213,   230,   230,   247,   247,   264,   264,   281,   281,
     298,   298,   315,   315,   332,   332,   349,   349,   366,   366,
     383,   383,   400,   400,   417,   417,   434,   434,   451,   451,
     468,   468,   485,   485,   502,   502,   519,   519,   536,   536,
     553,   553,   570,   570,   587,   587,   604,   605,   608,   609,
     610,   613,   650,   661,   662,   665,   672,   679,   688,   702,
     703,   710,   710,   723,   727,   728,   732,   737,   743,   747,
     751,   755,   761,   767,   773,   778,   782,   786,   792,   798,
     802,   806,   812,   816,   822,   823,   827,   833,   842,   848,
     866,   871,   883,   899,   905,   913,   933,   951,   960,   979,
     978,   993,   992,  1024,  1027,  1034,  1033,  1044,  1050,  1057,
    1064,  1075,  1081,  1084,  1092,  1091,  1102,  1108,  1120,  1124,
    1129,  1119,  1150,  1149,  1162,  1165,  1171,  1174,  1186,  1190,
    1185,  1208,  1207,  1223,  1224,  1228,  1232,  1236,  1240,  1244,
    1248,  1252,  1256,  1260,  1264,  1268,  1272,  1276,  1280,  1284,
    1288,  1292,  1296,  1301,  1307,  1308,  1312,  1323,  1327,  1331,
    1335,  1340,  1344,  1354,  1358,  1363,  1371,  1375,  1376,  1387,
    1391,  1395,  1399,  1403,  1411,  1412,  1418,  1425,  1431,  1438,
    1441,  1448,  1454,  1471,  1478,  1479,  1486,  1487,  1506,  1507,
    1510,  1513,  1517,  1528,  1537,  1543,  1546,  1549,  1556,  1557,
    1563,  1576,  1591,  1599,  1611,  1616,  1622,  1623,  1624,  1625,
    1626,  1627,  1633,  1634,  1635,  1636,  1642,  1643,  1644,  1645,
    1646,  1652,  1653,  1656,  1659,  1660,  1661,  1662,  1663,  1666,
    1667,  1680,  1684,  1689,  1694,  1699,  1703,  1704,  1707,  1713,
    1720,  1726,  1733,  1739,  1750,  1766,  1795,  1833,  1858,  1876,
    1885,  1888,  1896,  1900,  1904,  1911,  1917,  1922,  1934,  1937,
    1949,  1950,  1956,  1957,  1963,  1967,  1973,  1974,  1980,  1984,
    1990,  2013,  2018,  2024,  2030,  2037,  2046,  2055,  2070,  2076,
    2081,  2085,  2092,  2105,  2106,  2112,  2118,  2121,  2125,  2131,
    2134,  2143,  2146,  2147,  2151,  2152,  2158,  2159,  2160,  2161,
    2162,  2164,  2163,  2178,  2184,  2188,  2192,  2196,  2200,  2205,
    2224,  2230,  2238,  2242,  2248,  2252,  2258,  2262,  2268,  2272,
    2281,  2285,  2289,  2293,  2299,  2302,  2310,  2311,  2313,  2314,
    2317,  2320,  2323,  2326,  2329,  2332,  2335,  2338,  2341,  2344,
    2347,  2350,  2353,  2356,  2362,  2366,  2370,  2374,  2378,  2382,
    2402,  2409,  2420,  2421,  2422,  2425,  2426,  2429,  2433,  2443,
    2447,  2451,  2455,  2459,  2463,  2467,  2473,  2479,  2487,  2495,
    2501,  2508,  2524,  2546,  2550,  2556,  2559,  2562,  2566,  2576,
    2580,  2599,  2607,  2608,  2620,  2621,  2624,  2628,  2634,  2638,
    2644,  2648
};
#endif

#if YYDEBUG || YYERROR_VERBOSE || YYTOKEN_TABLE
/* YYTNAME[SYMBOL-NUM] -- String name of the symbol SYMBOL-NUM.
   First, the terminals, then, starting at YYNTOKENS, nonterminals.  */
const char *yytname[] =
{
  "$end", "error", "$undefined", "LLITERAL", "LASOP", "LCOLAS", "LBREAK",
  "LCASE", "LCHAN", "LCONST", "LCONTINUE", "LDDD", "LDEFAULT", "LDEFER",
  "LELSE", "LFALL", "LFOR", "LFUNC", "LGO", "LGOTO", "LIF", "LIMPORT",
  "LINTERFACE", "LMAP", "LNAME", "LPACKAGE", "LRANGE", "LRETURN",
  "LSELECT", "LSTRUCT", "LSWITCH", "LTYPE", "LVAR", "LANDAND", "LANDNOT",
  "LBODY", "LCOMM", "LDEC", "LEQ", "LGE", "LGT", "LIGNORE", "LINC", "LLE",
  "LLSH", "LLT", "LNE", "LOROR", "LRSH", "'+'", "'-'", "'|'", "'^'", "'*'",
  "'/'", "'%'", "'&'", "NotPackage", "NotParen", "'('", "')'",
  "PreferToRightParen", "';'", "'.'", "'$'", "'='", "':'", "'{'", "'}'",
  "'!'", "'~'", "'['", "']'", "'?'", "'@'", "','", "$accept", "file",
  "package", "loadsys", "loadcore", "@1", "loadlock", "@2", "loadsched",
  "@3", "loadsem", "@4", "loadgc", "@5", "loadprof", "@6", "loadchannels",
  "@7", "loadhash", "@8", "loadheapdump", "@9", "loadmaps", "@10",
  "loadnetpoll", "@11", "loadifacestuff", "@12", "loadvdso", "@13",
  "loadprintf", "@14", "loadstrings", "@15", "loadfp", "@16",
  "loadschedinit", "@17", "loadfinalize", "@18", "loadcgo", "@19",
  "loadsync", "@20", "loadcheck", "@21", "loadstackwb", "@22",
  "loaddefers", "@23", "loadseq", "@24", "loadruntime", "@25", "imports",
  "import", "import_stmt", "import_stmt_list", "import_here",
  "import_package", "import_safety", "import_there", "@26", "xdcl",
  "common_dcl", "lconst", "vardcl", "constdcl", "constdcl1", "typedclname",
  "typedcl", "simple_stmt", "case", "compound_stmt", "@27", "caseblock",
  "@28", "caseblock_list", "loop_body", "@29", "range_stmt", "for_header",
  "for_body", "for_stmt", "@30", "if_header", "if_stmt", "@31", "@32",
  "@33", "elseif", "@34", "elseif_list", "else", "switch_stmt", "@35",
  "@36", "select_stmt", "@37", "expr", "uexpr", "pseudocall",
  "pexpr_no_paren", "start_complit", "keyval", "bare_complitexpr",
  "complitexpr", "pexpr", "expr_or_type", "name_or_type", "lbrace",
  "new_name", "dcl_name", "onew_name", "sym", "hidden_importsym", "name",
  "labelname", "dotdotdot", "ntype", "non_expr_type", "non_recvchantype",
  "convtype", "comptype", "fnret_type", "dotname", "othertype", "ptrtype",
  "recvchantype", "structtype", "interfacetype", "xfndcl", "fndcl",
  "hidden_fndcl", "fntype", "fnbody", "fnres", "fnlitdcl", "fnliteral",
  "xdcl_list", "vardcl_list", "constdcl_list", "typedcl_list",
  "structdcl_list", "interfacedcl_list", "structdcl", "packname", "embed",
  "interfacedcl", "indcl", "arg_type", "arg_type_list",
  "oarg_type_list_ocomma", "stmt", "non_dcl_stmt", "@38", "stmt_list",
  "new_name_list", "dcl_name_list", "expr_list", "expr_or_type_list",
  "keyval_list", "braced_keyval_list", "osemi", "ocomma", "oexpr",
  "oexpr_list", "osimple_stmt", "ohidden_funarg_list",
  "ohidden_structdcl_list", "ohidden_interfacedcl_list", "oliteral",
  "hidden_import", "hidden_pkg_importsym", "hidden_pkgtype", "hidden_type",
  "hidden_type_non_recv_chan", "hidden_type_misc", "hidden_type_recv_chan",
  "hidden_type_func", "hidden_funarg", "hidden_structdcl",
  "hidden_interfacedcl", "ohidden_funres", "hidden_funres",
  "hidden_literal", "hidden_constant", "hidden_import_list",
  "hidden_funarg_list", "hidden_structdcl_list",
  "hidden_interfacedcl_list", 0
};
#endif

# ifdef YYPRINT
/* YYTOKNUM[YYLEX-NUM] -- Internal token number corresponding to
   token YYLEX-NUM.  */
static const yytype_uint16 yytoknum[] =
{
       0,   256,   257,   258,   259,   260,   261,   262,   263,   264,
     265,   266,   267,   268,   269,   270,   271,   272,   273,   274,
     275,   276,   277,   278,   279,   280,   281,   282,   283,   284,
     285,   286,   287,   288,   289,   290,   291,   292,   293,   294,
     295,   296,   297,   298,   299,   300,   301,   302,   303,    43,
      45,   124,    94,    42,    47,    37,    38,   304,   305,    40,
      41,   306,    59,    46,    36,    61,    58,   123,   125,    33,
     126,    91,    93,    63,    64,    44
};
# endif

/* YYR1[YYN] -- Symbol number of symbol that rule YYN derives.  */
static const yytype_uint16 yyr1[] =
{
       0,    76,    77,    78,    78,    79,    81,    80,    83,    82,
      85,    84,    87,    86,    89,    88,    91,    90,    93,    92,
      95,    94,    97,    96,    99,    98,   101,   100,   103,   102,
     105,   104,   107,   106,   109,   108,   111,   110,   113,   112,
     115,   114,   117,   116,   119,   118,   121,   120,   123,   122,
     125,   124,   127,   126,   129,   128,   130,   130,   131,   131,
     131,   132,   132,   133,   133,   134,   134,   134,   135,   136,
     136,   138,   137,   139,   139,   139,   139,   139,   140,   140,
     140,   140,   140,   140,   140,   140,   140,   140,   141,   142,
     142,   142,   143,   143,   144,   144,   144,   145,   146,   147,
     147,   147,   147,   147,   147,   148,   148,   148,   148,   150,
     149,   152,   151,   153,   153,   155,   154,   156,   156,   156,
     157,   157,   157,   158,   160,   159,   161,   161,   163,   164,
     165,   162,   167,   166,   168,   168,   169,   169,   171,   172,
     170,   174,   173,   175,   175,   175,   175,   175,   175,   175,
     175,   175,   175,   175,   175,   175,   175,   175,   175,   175,
     175,   175,   175,   175,   176,   176,   176,   176,   176,   176,
     176,   176,   176,   177,   177,   177,   178,   178,   178,   178,
     178,   178,   178,   178,   178,   178,   178,   178,   178,   178,
     179,   180,   181,   181,   182,   182,   183,   183,   184,   184,
     185,   186,   186,   187,   188,   189,   189,   190,   190,   190,
     191,   191,   192,   193,   194,   194,   195,   195,   195,   195,
     195,   195,   196,   196,   196,   196,   197,   197,   197,   197,
     197,   198,   198,   199,   200,   200,   200,   200,   200,   201,
     201,   202,   202,   202,   202,   202,   202,   202,   203,   204,
     205,   205,   206,   206,   207,   208,   208,   209,   209,   210,
     211,   211,   212,   212,   212,   213,   214,   214,   215,   215,
     216,   216,   217,   217,   218,   218,   219,   219,   220,   220,
     221,   221,   221,   221,   221,   221,   222,   222,   223,   224,
     224,   224,   225,   226,   226,   226,   226,   227,   227,   228,
     228,   229,   229,   229,   229,   229,   230,   230,   230,   230,
     230,   231,   230,   230,   230,   230,   230,   230,   230,   230,
     232,   232,   233,   233,   234,   234,   235,   235,   236,   236,
     237,   237,   237,   237,   238,   238,   239,   239,   240,   240,
     241,   241,   242,   242,   243,   243,   244,   244,   245,   245,
     246,   246,   247,   247,   248,   248,   248,   248,   248,   248,
     249,   250,   251,   251,   251,   252,   252,   253,   253,   253,
     253,   253,   253,   253,   253,   253,   253,   253,   254,   255,
     256,   256,   257,   258,   258,   259,   259,   260,   260,   261,
     261,   261,   262,   262,   263,   263,   264,   264,   265,   265,
     266,   266
};

/* YYR2[YYN] -- Number of symbols composing right hand side of rule YYN.  */
static const yytype_uint8 yyr2[] =
{
       0,     2,     4,     0,     3,    25,     0,     3,     0,     3,
       0,     3,     0,     3,     0,     3,     0,     3,     0,     3,
       0,     3,     0,     3,     0,     3,     0,     3,     0,     3,
       0,     3,     0,     3,     0,     3,     0,     3,     0,     3,
       0,     3,     0,     3,     0,     3,     0,     3,     0,     3,
       0,     3,     0,     3,     0,     3,     0,     3,     2,     5,
       3,     3,     2,     1,     3,     1,     2,     2,     4,     0,
       1,     0,     4,     0,     1,     1,     1,     1,     2,     5,
       3,     2,     5,     7,     3,     2,     5,     3,     1,     2,
       4,     3,     4,     3,     1,     2,     1,     1,     2,     1,
       3,     3,     3,     2,     2,     3,     5,     5,     2,     0,
       4,     0,     3,     0,     2,     0,     4,     4,     4,     2,
       5,     1,     1,     2,     0,     3,     1,     3,     0,     0,
       0,     8,     0,     5,     0,     2,     0,     2,     0,     0,
       7,     0,     5,     1,     3,     3,     3,     3,     3,     3,
       3,     3,     3,     3,     3,     3,     3,     3,     3,     3,
       3,     3,     3,     3,     1,     2,     2,     2,     2,     2,
       2,     2,     2,     3,     5,     6,     1,     1,     3,     5,
       5,     4,     6,     8,     1,     5,     5,     5,     7,     1,
       0,     3,     1,     4,     1,     4,     1,     3,     1,     1,
       1,     1,     1,     1,     1,     0,     1,     1,     1,     1,
       4,     4,     1,     1,     1,     2,     1,     1,     1,     1,
       1,     3,     1,     1,     1,     2,     1,     1,     1,     1,
       3,     1,     1,     1,     1,     1,     1,     1,     1,     1,
       3,     4,     4,     2,     3,     5,     1,     1,     2,     3,
       5,     3,     5,     3,     3,     5,     8,     5,     8,     5,
       0,     3,     0,     1,     3,     1,     4,     2,     0,     3,
       1,     3,     1,     3,     1,     3,     1,     3,     1,     3,
       3,     2,     4,     3,     5,     5,     1,     3,     1,     2,
       1,     3,     4,     1,     2,     2,     1,     1,     3,     0,
       2,     0,     1,     1,     1,     1,     1,     1,     1,     1,
       1,     0,     4,     1,     2,     2,     2,     2,     2,     2,
       1,     3,     1,     3,     1,     3,     1,     3,     1,     3,
       1,     1,     3,     3,     0,     2,     0,     1,     0,     1,
       0,     1,     0,     1,     0,     1,     0,     1,     0,     1,
       0,     1,     0,     1,     4,     4,     5,     6,     4,     4,
       1,     1,     1,     1,     1,     1,     1,     1,     1,     3,
       4,     5,     4,     4,     2,     2,     4,     3,     3,     5,
       3,     4,     3,     5,     1,     0,     1,     3,     1,     1,
       2,     1,     1,     5,     0,     2,     1,     3,     1,     3,
       1,     3
};

/* YYDEFACT[STATE-NAME] -- Default rule to reduce with in state
   STATE-NUM when YYTABLE doesn't specify something else to do.  Zero
   means the default is an error.  */
static const yytype_uint16 yydefact[] =
{
       6,     0,     3,     8,     0,     1,     0,    56,    10,     0,
       0,    71,   207,   209,     0,     0,   208,   268,    12,     0,
      71,    69,     7,   394,     0,     4,     0,     0,     0,    14,
       0,    71,     9,    70,     0,     0,     0,    65,     0,     0,
      58,    71,     0,    57,    77,   176,   205,     0,    88,   205,
       0,   313,   124,     0,     0,     0,   128,     0,     0,   342,
     141,     0,   138,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,   340,     0,    74,     0,   306,   307,
     310,   308,   309,    99,   143,   184,   196,   164,   213,   212,
     177,     0,     0,     0,   233,   246,   247,    75,   265,     0,
     189,    76,     0,    16,     0,    71,    11,    68,     0,     0,
       0,     0,     0,     0,   395,   210,   211,    60,    63,   336,
      67,    71,    62,    66,   206,   314,   203,     0,     0,     0,
       0,   212,   239,   243,   229,   227,   228,   226,   315,   184,
       0,   344,   299,     0,   260,   184,   318,   344,   201,   202,
       0,     0,   326,   343,   319,     0,     0,   344,     0,     0,
      85,    97,     0,    78,   324,   204,     0,   172,   167,   168,
     171,   165,   166,     0,     0,   198,     0,   199,   224,   222,
     223,   169,   170,     0,   341,     0,   269,     0,    81,     0,
       0,     0,     0,     0,   104,     0,     0,     0,   103,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,   190,     0,     0,   340,   311,     0,   190,
     267,     0,     0,     0,     0,    18,     0,    71,    13,   360,
       0,     0,   260,     0,     0,   361,     0,     0,    72,   337,
       0,    61,   299,     0,     0,   244,   220,   218,   219,   216,
     217,   248,     0,     0,     0,   345,   122,     0,   125,     0,
     121,   214,   293,   212,   296,   200,   297,   338,     0,   299,
       0,   254,   129,   126,   207,     0,   253,     0,   336,   290,
     278,     0,   113,     0,     0,   251,   322,   336,   276,   288,
     352,     0,   139,    87,   274,   336,    98,    80,   270,   336,
       0,     0,    89,     0,   225,   197,     0,     0,    84,   336,
       0,     0,   100,   145,   160,   163,   146,   150,   151,   149,
     161,   148,   147,   144,   162,   152,   153,   154,   155,   156,
     157,   158,   159,   334,   173,   328,   338,     0,   178,   341,
       0,     0,   338,   334,   305,   109,   303,   302,   320,   304,
       0,   102,   101,   327,    20,     0,    71,    15,     0,     0,
       0,     0,   368,     0,     0,     0,     0,     0,   367,     0,
     362,   363,   364,     0,   396,     0,     0,   346,     0,     0,
       0,    64,    59,     0,     0,     0,   230,   240,   119,   115,
     123,     0,     0,   344,   215,   294,   295,   339,   300,   262,
       0,     0,     0,   344,     0,   286,     0,   299,   289,   337,
       0,     0,     0,     0,   352,     0,     0,   337,     0,   353,
     281,     0,   352,     0,   337,     0,   337,     0,    91,   325,
       0,     0,     0,   249,   220,   218,   219,   217,   190,   242,
     241,   337,     0,    93,     0,   190,   192,   330,   331,   338,
       0,   338,   339,     0,     0,     0,   181,   340,   312,   339,
       0,     0,     0,     0,   266,    22,     0,    71,    17,     0,
       0,   375,   365,   366,   346,   350,     0,   348,     0,   374,
     389,     0,     0,   391,   392,     0,     0,     0,     0,     0,
     352,     0,     0,   359,     0,   347,   354,   358,   355,   262,
     221,     0,     0,     0,     0,   298,   299,   212,   263,   238,
     236,   237,   234,   235,   259,   262,   261,   130,   127,   287,
     291,     0,   279,   252,   245,     0,     0,   142,   111,   114,
       0,   283,     0,   352,   277,   250,   323,   280,   113,   275,
      86,   271,    79,    90,     0,   334,    94,   272,   336,    96,
      82,    92,   334,     0,   339,   335,   187,     0,   329,   174,
     180,   179,     0,   185,   186,     0,   321,    24,     0,    71,
      19,   377,     0,     0,   368,     0,   367,     0,   384,   400,
     351,     0,     0,     0,   398,   349,   378,   390,     0,   356,
       0,   369,     0,   352,   380,     0,   397,   385,     0,   118,
     117,   344,     0,   299,   255,   134,   262,     0,   108,     0,
     352,   352,   282,     0,   221,     0,   337,     0,    95,     0,
     190,   194,   191,   332,   333,   175,   340,   182,   110,    26,
       0,    71,    21,   376,   385,   346,   373,     0,     0,   352,
     372,     0,     0,   370,   357,   381,   346,   346,   388,   257,
     386,   116,   120,   264,     0,   136,   292,     0,     0,   105,
       0,   112,   285,   284,   140,   188,   273,    83,   193,   334,
       0,    28,     0,    71,    23,   379,     0,   401,   371,   382,
     399,     0,     0,     0,   262,     0,   135,   131,     0,     0,
       0,   183,    30,     0,    71,    25,   385,   393,   385,   387,
     256,   132,   137,   107,   106,   195,    32,     0,    71,    27,
     383,   258,   344,    34,     0,    71,    29,     0,    36,     0,
      71,    31,   133,    38,     0,    71,    33,    40,     0,    71,
      35,    42,     0,    71,    37,    44,     0,    71,    39,    46,
       0,    71,    41,    48,     0,    71,    43,    50,     0,    71,
      45,    52,     0,    71,    47,    54,     0,    71,    49,     5,
       0,    71,    51,    71,    53,    55
};

/* YYDEFGOTO[NTERM-NUM].  */
static const yytype_int16 yydefgoto[] =
{
      -1,     1,     7,     2,     3,     4,     8,     9,    18,    19,
      29,    30,   103,   104,   225,   226,   354,   355,   465,   466,
     567,   568,   629,   630,   671,   672,   692,   693,   706,   707,
     713,   714,   718,   719,   723,   724,   727,   728,   731,   732,
     735,   736,   739,   740,   743,   744,   747,   748,   751,   752,
     755,   756,   759,   760,    17,    27,    40,   119,    41,    11,
      34,    22,    23,    75,   346,    77,   163,   546,   547,   159,
     160,    78,   528,   347,   462,   529,   609,   412,   390,   501,
     256,   257,   258,    79,   141,   272,    80,   147,   402,   605,
     686,   712,   655,   687,    81,   157,   423,    82,   155,    83,
      84,    85,    86,   333,   447,   448,   622,    87,   335,   262,
     150,    88,   164,   125,   131,    16,    90,    91,   264,   265,
     177,   133,    92,    93,   508,   246,    94,   248,   249,    95,
      96,    97,   144,   232,    98,   271,   514,    99,   100,    28,
     299,   548,   295,   287,   278,   288,   289,   290,   280,   408,
     266,   267,   268,   348,   349,   341,   350,   291,   166,   102,
     336,   449,   450,   240,   398,   185,   154,   273,   494,   583,
     577,   420,   114,   230,   236,   648,   471,   370,   371,   372,
     374,   584,   579,   649,   650,   484,   485,    35,   495,   585,
     580
};

/* YYPACT[STATE-NUM] -- Index in YYTABLE of the portion describing
   STATE-NUM.  */
#define YYPACT_NINF -545
static const yytype_int16 yypact[] =
{
    -545,    48,    43,  -545,    46,  -545,    56,  -545,  -545,    46,
      57,  -545,  -545,  -545,    55,    26,  -545,    84,  -545,    46,
    -545,   114,  -545,  -545,    97,  -545,    38,   117,  1163,  -545,
      46,  -545,  -545,  -545,   128,   301,    30,  -545,    62,   190,
    -545,    46,   207,  -545,  -545,  -545,    56,  1987,  -545,    56,
     493,  -545,  -545,   150,   493,    56,  -545,   116,   149,  1833,
    -545,   116,  -545,   182,   285,  1833,  1833,  1833,  1833,  1833,
    1833,  1876,  1833,  1833,   789,   160,  -545,   351,  -545,  -545,
    -545,  -545,  -545,  1490,  -545,  -545,   171,    28,  -545,   179,
    -545,   194,   204,   116,   228,  -545,  -545,  -545,   231,    81,
    -545,  -545,    91,  -545,    46,  -545,  -545,  -545,   222,    20,
     273,   222,   222,   242,  -545,  -545,  -545,  -545,  -545,   237,
    -545,  -545,  -545,  -545,  -545,  -545,  -545,   256,  1995,  1995,
    1995,  -545,   248,  -545,  -545,  -545,  -545,  -545,  -545,   252,
      28,   908,   749,   258,   260,   266,  -545,  1833,  -545,  -545,
     267,  1995,  2256,   244,  -545,   294,   546,  1833,   343,  1995,
    -545,  -545,   389,  -545,  -545,  -545,   510,  -545,  -545,  -545,
    -545,  -545,  -545,  1919,  1876,  2256,   276,  -545,   111,  -545,
     184,  -545,  -545,   265,  2256,   275,  -545,   432,  -545,  1962,
    1833,  1833,  1833,  1833,  -545,  1833,  1833,  1833,  -545,  1833,
    1833,  1833,  1833,  1833,  1833,  1833,  1833,  1833,  1833,  1833,
    1833,  1833,  1833,  -545,  1001,   506,  1833,  -545,  1833,  -545,
    -545,  1412,  1833,  1833,  1833,  -545,    46,  -545,  -545,  -545,
     978,    56,   260,   286,   348,  -545,  1691,  1691,  -545,    95,
     293,  -545,   749,   347,  1995,  -545,  -545,  -545,  -545,  -545,
    -545,  -545,   302,    56,  1833,  -545,  -545,   319,  -545,   102,
     306,  1995,  -545,   749,  -545,  -545,  -545,   296,   318,   749,
    1412,  -545,  -545,   321,   186,   357,  -545,   329,   327,  -545,
    -545,   328,  -545,    51,   147,  -545,  -545,   335,  -545,  -545,
     409,   844,  -545,  -545,  -545,   360,  -545,  -545,  -545,   364,
    1833,    56,   363,  2021,  -545,   365,  1995,  1995,  -545,   368,
    1833,   370,  2256,  2327,  -545,  2280,   617,   617,   617,   617,
    -545,   617,   617,  2304,  -545,   773,   773,   773,   773,  -545,
    -545,  -545,  -545,  1545,  -545,  -545,    34,  1622,  -545,  2154,
     371,  1317,  2121,  1545,  -545,  -545,  -545,  -545,  -545,  -545,
     119,   244,   244,  2256,  -545,    46,  -545,  -545,  1061,   380,
     373,   372,  -545,   374,   436,  1691,    54,    73,  -545,   381,
    -545,  -545,  -545,  1069,  -545,   122,   386,    56,   388,   390,
     391,  -545,  -545,   405,  1995,   406,  -545,  -545,  2256,  -545,
    -545,  1680,  1735,  1833,  -545,  -545,  -545,   749,  -545,  2048,
     407,   169,   319,  1833,    56,   410,   415,   749,  -545,   513,
     383,  1995,    65,   357,   409,   357,   416,   346,   401,  -545,
    -545,    56,   409,   435,    56,   418,    56,   421,   244,  -545,
    1833,  2074,  1995,  -545,   212,   254,   289,   290,  -545,  -545,
    -545,    56,   423,   244,  1833,  -545,  2184,  -545,  -545,   412,
     422,   420,  1876,   424,   433,   437,  -545,  1833,  -545,  -545,
     439,   434,  1412,  1317,  -545,  -545,    46,  -545,  -545,  1691,
     464,  -545,  -545,  -545,    56,  1555,  1691,    56,  1691,  -545,
    -545,   504,    99,  -545,  -545,   446,   447,  1691,    54,  1691,
     409,    56,    56,  -545,   449,   445,  -545,  -545,  -545,  2048,
    -545,  1412,  1833,  1833,   459,  -545,   749,   452,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,  2048,  -545,  -545,  -545,  -545,
    -545,   463,  -545,  -545,  -545,  1876,   458,  -545,  -545,  -545,
     465,  -545,   468,   409,  -545,  -545,  -545,  -545,  -545,  -545,
    -545,  -545,  -545,   244,   469,  1545,  -545,  -545,   474,  1962,
    -545,   244,  1545,  1778,  1545,  -545,  -545,   471,  -545,  -545,
    -545,  -545,    52,  -545,  -545,   224,  -545,  -545,    46,  -545,
    -545,  -545,   478,   485,   488,   492,   494,   486,  -545,  -545,
     498,   489,  1691,   508,  -545,   500,  -545,  -545,   525,  -545,
    1691,  -545,   515,   409,  -545,   519,  -545,  2082,   226,  2256,
    2256,  1833,   522,   749,  -545,  -545,  2048,   138,  -545,  1317,
     409,   409,  -545,   189,   325,   520,    56,   529,   370,   523,
    -545,  2256,  -545,  -545,  -545,  -545,  1833,  -545,  -545,  -545,
      46,  -545,  -545,  -545,  2082,    56,  -545,  1555,  1691,   409,
    -545,    56,    99,  -545,  -545,  -545,    56,    56,  -545,  -545,
    -545,  -545,  -545,  -545,   530,   580,  -545,  1833,  1833,  -545,
    1876,   535,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  1545,
     532,  -545,    46,  -545,  -545,  -545,   541,  -545,  -545,  -545,
    -545,   547,   548,   549,  2048,    33,  -545,  -545,  2208,  2232,
     542,  -545,  -545,    46,  -545,  -545,  2082,  -545,  2082,  -545,
    -545,  -545,  -545,  -545,  -545,  -545,  -545,    46,  -545,  -545,
    -545,  -545,  1833,  -545,    46,  -545,  -545,   319,  -545,    46,
    -545,  -545,  -545,  -545,    46,  -545,  -545,  -545,    46,  -545,
    -545,  -545,    46,  -545,  -545,  -545,    46,  -545,  -545,  -545,
      46,  -545,  -545,  -545,    46,  -545,  -545,  -545,    46,  -545,
    -545,  -545,    46,  -545,  -545,  -545,    46,  -545,  -545,  -545,
      46,  -545,  -545,  -545,  -545,  -545
};

/* YYPGOTO[NTERM-NUM].  */
static const yytype_int16 yypgoto[] =
{
    -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,  -545,   -18,  -545,  -545,    -9,
    -545,   -13,  -545,  -545,   583,  -545,  -154,   -37,    -4,  -545,
    -142,  -110,  -545,   -70,  -545,  -545,  -545,    78,  -387,  -545,
    -545,  -545,  -545,  -545,  -545,  -155,  -545,  -545,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  -545,  1068,
     235,    39,  -545,  -213,    63,    68,  -545,   161,   -67,   361,
     141,   -19,   322,   576,    -3,   -42,   705,  -545,   366,   -45,
     456,  -545,  -545,  -545,  -545,   -38,   384,   -35,   -57,  -545,
    -545,  -545,  -545,  -545,   635,   399,  -486,  -545,  -545,  -545,
    -545,  -545,  -545,  -545,  -545,   216,  -133,  -250,   227,  -545,
     238,  -545,  -231,  -302,   609,  -545,  -257,  -545,   -72,   -15,
     115,  -545,  -319,  -236,  -287,  -215,  -545,  -119,  -455,  -545,
    -545,  -358,  -545,   118,  -545,   -96,  -545,   283,   172,   292,
     156,    11,    18,  -544,  -545,  -456,   170,  -545,   426,  -545,
    -545
};

/* YYTABLE[YYPACT[STATE-NUM]].  What to do in state STATE-NUM.  If
   positive, shift that token.  If negative, reduce the rule which
   number is the opposite.  If zero, do what YYDEFACT says.
   If YYTABLE_NINF, syntax error.  */
#define YYTABLE_NINF -327
static const yytype_int16 yytable[] =
{
      20,   340,   292,    15,   176,   189,   343,    32,   298,   134,
      31,   383,   136,   401,   179,   517,   294,   279,   106,   573,
     118,   105,   260,    42,   461,    89,   588,   124,   122,   604,
     124,   255,   121,   414,   416,    42,   146,   255,   400,   458,
     188,    37,   410,   126,   153,   451,   126,   255,     5,   453,
     143,   418,   126,   701,   115,   460,   531,   480,    24,   425,
     161,   165,    12,   427,   537,    37,   229,   229,     6,   229,
     229,    10,   525,   442,   165,   405,   486,   526,    12,   231,
      12,    21,   220,   245,   251,   252,    12,   214,    25,   139,
     675,   215,   228,   145,    14,   227,   222,    38,    37,   216,
     345,    39,   480,   116,   481,    26,   281,   391,   241,   452,
     413,    13,    14,   482,   296,   189,   148,   179,   626,    12,
     656,   302,   117,    12,   627,    39,   259,    13,    14,    13,
      14,   277,   594,   527,   369,    13,    14,   286,    33,   263,
     379,   380,   406,   657,   311,   487,  -233,   126,   149,   481,
     309,   148,   710,   126,   711,   161,   223,   179,    39,   165,
      36,   566,   555,   530,   557,   532,   224,   392,    13,    14,
    -232,   405,    13,    14,    12,   612,   521,   224,  -233,    43,
     676,   463,   491,   149,   165,  -265,   681,   464,   368,  -286,
     107,   682,   683,   120,   368,   368,   525,   492,   700,   385,
     415,   526,   156,   658,   659,   565,    12,   351,   352,   142,
     123,   140,   338,   660,   357,   140,   394,   356,    89,  -265,
     151,   381,   186,    13,    14,   545,   615,   233,   373,   235,
     237,   463,   552,   619,   219,   645,    42,   516,   213,   263,
     221,   158,   562,  -231,   598,  -203,   422,  -229,  -286,   404,
     387,  -265,   662,   663,  -286,    13,    14,   664,   433,  -317,
     217,   439,   440,   218,  -317,   434,   263,    89,   436,   479,
     455,  -229,   541,  -316,   504,   602,   279,   490,  -316,  -229,
     179,   679,   539,   255,   518,   428,   463,  -232,   463,  -227,
    -231,   274,   628,   255,   651,   443,    14,   234,   165,   239,
     167,   168,   169,   170,   171,   172,   238,   181,   182,    12,
     108,   253,   617,  -227,  -317,   242,   368,   269,   109,   224,
    -317,  -227,   110,   368,  -228,  -226,   275,   270,  -316,   282,
     722,   368,   111,   112,  -316,   276,   305,   306,    89,   433,
      13,    14,   512,   468,   162,   377,   467,   307,  -228,  -226,
     690,   378,   661,   382,   389,   384,  -228,  -226,    13,    14,
    -230,   509,   386,   483,   511,   113,   524,    12,   393,   549,
     274,   397,   654,   571,   373,    12,   351,   352,   399,   578,
     581,   405,   586,   403,  -230,   558,   245,   544,   407,   409,
     277,   591,  -230,   593,   263,   179,   507,   417,   286,   283,
     411,   519,   536,   293,   263,   284,   126,   669,   167,   171,
     187,   670,   419,    12,   126,   543,    13,    14,   126,    13,
      14,   161,   424,   165,    13,    14,   426,   368,   430,   551,
     441,   135,   438,   576,   368,   444,   368,   457,   165,   474,
     475,   477,   512,   476,   478,   368,   488,   368,   493,   297,
     496,   523,   497,   498,   570,   178,    12,   569,   512,    89,
      89,   509,    13,    14,   511,   499,   500,   515,   179,   535,
     538,   373,   575,   404,   582,   520,   533,   509,   540,   483,
     511,   542,   652,   550,   559,   483,   639,   554,   595,   373,
     556,   255,   308,   560,   643,   459,    45,   561,    89,   563,
     364,    47,   564,   263,   618,    13,    14,   587,   589,   597,
     127,   603,   247,   247,   247,    57,    58,    12,    47,   590,
     492,   601,    61,   606,   608,   610,   247,   127,   611,   614,
      12,   625,    57,    58,    12,   247,   616,   274,   633,    61,
     368,   578,   678,   247,   549,   634,   243,  -207,   368,   512,
     247,   635,    71,  -208,   636,   368,   632,   717,   178,   631,
     637,   638,   641,   129,    74,   337,    13,    14,   509,   244,
     274,   511,   275,   247,   642,   300,   640,   644,   646,    13,
      14,    74,   653,    13,    14,   301,    13,    14,   665,   667,
     684,   668,   368,   558,   685,   576,   368,   463,   178,   283,
     263,   696,   255,   179,   691,   284,    89,   697,   698,   699,
     705,    76,   666,   165,   285,   702,   613,   623,   674,    13,
      14,   673,   624,   429,   395,   138,   247,   512,   247,   396,
     304,   376,   373,   534,   575,   505,   522,   101,   582,   483,
     607,   472,   572,   373,   373,   247,   509,   247,   596,   511,
     473,   192,   680,   247,   368,   677,   368,   375,   592,     0,
     695,   200,     0,   694,     0,   204,   205,   206,   207,   208,
     209,   210,   211,   212,     0,   247,     0,     0,     0,     0,
       0,   709,   137,     0,   708,     0,     0,   435,     0,     0,
     247,   247,     0,     0,     0,   716,     0,     0,   715,     0,
       0,     0,   721,     0,     0,   720,   180,   726,     0,     0,
     725,     0,   730,     0,     0,   729,   734,     0,     0,   733,
     738,   178,     0,   737,   742,     0,     0,   741,   746,     0,
       0,   745,   750,     0,     0,   749,   754,     0,     0,   753,
     758,     0,     0,   757,   762,     0,     0,   761,   764,     0,
     765,   763,   132,     0,     0,     0,     0,    47,     0,     0,
     261,     0,     0,   250,   250,   250,   127,     0,   247,     0,
       0,    57,    58,    12,     0,     0,     0,   250,    61,     0,
       0,   247,     0,   510,     0,   243,   250,     0,     0,     0,
       0,   247,    45,     0,   250,   247,     0,    47,     0,     0,
     183,   250,   129,     0,     0,     0,   127,   192,   244,   180,
       0,    57,    58,    12,     0,   247,   247,   200,    61,     0,
      74,   204,    13,    14,   250,    65,   209,   210,   211,   212,
       0,     0,     0,   132,   132,   132,   178,     0,    66,    67,
       0,    68,    69,     0,     0,    70,     0,   132,    71,   180,
       0,     0,    47,     0,     0,     0,   132,     0,    72,    73,
      74,   127,    13,    14,   132,     0,    57,    58,    12,     0,
       0,   132,     0,    61,     0,     0,     0,   250,     0,   250,
     243,     0,     0,   510,     0,     0,     0,     0,     0,     0,
     247,     0,     0,     0,   132,     0,   250,   129,   250,   510,
       0,     0,     0,   244,   250,     0,     0,     0,     0,   178,
       0,    45,     0,     0,     0,    74,    47,    13,    14,   421,
       0,     0,     0,     0,     0,   127,   250,     0,     0,     0,
      57,    58,    12,   247,   254,     0,     0,    61,   437,     0,
       0,   250,   250,     0,    65,     0,     0,   132,     0,   132,
       0,     0,     0,     0,     0,     0,     0,    66,    67,     0,
      68,    69,     0,     0,    70,     0,   132,    71,   132,     0,
       0,     0,   180,     0,   132,     0,     0,    72,    73,    74,
       0,    13,    14,     0,     0,     0,   358,   247,     0,     0,
     510,     0,     0,     0,     0,   359,   132,     0,     0,     0,
     360,   361,   362,     0,    45,     0,     0,   363,   132,    47,
       0,   132,   132,     0,   364,     0,     0,     0,   127,   250,
       0,     0,     0,    57,    58,    12,     0,     0,     0,     0,
      61,   365,   250,     0,   513,     0,     0,   173,     0,     0,
       0,     0,   250,   366,   178,     0,   250,     0,     0,   367,
      66,    67,    14,    68,   174,     0,     0,    70,     0,     0,
      71,   334,     0,     0,     0,     0,   250,   250,   510,   358,
      72,    73,    74,     0,    13,    14,     0,   358,   359,     0,
     489,     0,     0,   360,   361,   362,   359,   180,     0,   132,
     363,   360,   361,   362,     0,     0,     0,   469,   363,     0,
       0,     0,   132,     0,   132,   364,     0,     0,     0,     0,
       0,     0,   132,     0,   365,     0,   132,     0,     0,     0,
     470,     0,   365,     0,     0,     0,     0,   152,     0,     0,
       0,     0,   367,     0,   513,    14,   132,   132,     0,   175,
     367,   250,   184,    14,     0,     0,     0,     0,     0,     0,
     513,     0,     0,     0,     0,     0,     0,     0,     0,     0,
     180,     0,     0,    -2,    44,     0,    45,     0,     0,    46,
       0,    47,    48,    49,     0,     0,    50,     0,    51,    52,
      53,    54,    55,    56,   250,    57,    58,    12,     0,     0,
      59,    60,    61,    62,    63,    64,     0,     0,     0,    65,
       0,     0,     0,     0,   132,     0,     0,     0,     0,     0,
       0,   132,    66,    67,     0,    68,    69,     0,     0,    70,
     132,     0,    71,     0,     0,   -73,     0,     0,     0,     0,
       0,     0,    72,    73,    74,     0,    13,    14,   250,     0,
       0,   513,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,   132,     0,     0,     0,   312,   313,
     314,   315,     0,   316,   317,   318,     0,   319,   320,   321,
     322,   323,   324,   325,   326,   327,   328,   329,   330,   331,
     332,     0,   175,     0,   339,     0,   342,     0,     0,     0,
     152,   152,   353,     0,     0,   180,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,   132,     0,
       0,   132,     0,     0,     0,     0,     0,     0,   344,   513,
      45,     0,   388,    46,  -301,    47,    48,    49,     0,  -301,
      50,     0,    51,    52,   127,    54,    55,    56,     0,    57,
      58,    12,     0,     0,    59,    60,    61,    62,    63,    64,
       0,     0,     0,    65,     0,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,    66,    67,   152,    68,
      69,     0,     0,    70,     0,     0,    71,     0,   152,  -301,
       0,     0,     0,     0,   345,  -301,    72,    73,    74,   132,
      13,    14,     0,     0,     0,     0,     0,     0,     0,     0,
       0,   446,     0,     0,     0,   175,     0,     0,     0,     0,
       0,   446,     0,   344,     0,    45,     0,     0,    46,     0,
      47,    48,    49,     0,     0,    50,     0,    51,    52,   127,
      54,    55,    56,     0,    57,    58,    12,     0,     0,    59,
      60,    61,    62,    63,    64,     0,     0,     0,    65,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,   152,
     152,    66,    67,     0,    68,    69,     0,     0,    70,     0,
       0,    71,     0,     0,  -301,     0,     0,     0,     0,   345,
    -301,    72,    73,    74,     0,    13,    14,     0,     0,     0,
       0,     0,     0,     0,   190,  -326,     0,     0,   152,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,     0,   152,     0,     0,     0,     0,     0,     0,     0,
     175,     0,     0,   191,   192,   184,   193,   194,   195,   196,
     197,     0,   198,   199,   200,   201,   202,   203,   204,   205,
     206,   207,   208,   209,   210,   211,   212,     0,    45,     0,
       0,     0,     0,    47,     0,  -326,     0,     0,     0,     0,
       0,     0,   127,   358,     0,  -326,     0,    57,    58,    12,
     599,   600,   359,     0,    61,     0,     0,   360,   361,   574,
       0,    65,     0,     0,   363,     0,     0,     0,     0,     0,
       0,   364,     0,   175,    66,    67,     0,    68,    69,     0,
       0,    70,     0,     0,    71,     0,     0,     0,   365,     0,
       0,     0,   445,   446,    72,    73,    74,     0,    13,    14,
     446,   621,   446,     0,     0,    45,   367,     0,    13,    14,
      47,     0,     0,     0,     0,     0,     0,     0,     0,   127,
       0,     0,     0,     0,    57,    58,    12,     0,     0,     0,
       0,    61,     0,   454,     0,     0,     0,     0,   173,     0,
       0,     0,     0,     0,     0,     0,     0,     0,     0,     0,
       0,    66,    67,     0,    68,   174,     0,     0,    70,     0,
       0,    71,     0,    45,     0,     0,     0,     0,    47,     0,
       0,    72,    73,    74,   184,    13,    14,   127,     0,   358,
       0,     0,    57,    58,    12,     0,   502,     0,   359,    61,
       0,     0,     0,   360,   361,   362,    65,     0,     0,     0,
     363,     0,     0,     0,     0,   688,   689,   364,   175,    66,
      67,     0,    68,    69,     0,     0,    70,   446,    45,    71,
       0,     0,     0,    47,   365,     0,     0,     0,     0,    72,
      73,    74,   127,    13,    14,     0,     0,    57,    58,    12,
       0,   503,   367,     0,    61,    14,     0,     0,     0,     0,
       0,    65,     0,     0,     0,     0,     0,     0,     0,     0,
       0,    45,     0,     0,    66,    67,    47,    68,    69,     0,
       0,    70,     0,     0,    71,   127,     0,     0,     0,     0,
      57,    58,    12,     0,    72,    73,    74,    61,    13,    14,
       0,     0,     0,     0,    65,     0,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,    66,    67,     0,
      68,    69,     0,     0,    70,     0,    45,    71,     0,     0,
       0,    47,     0,     0,     0,   620,     0,    72,    73,    74,
     127,    13,    14,     0,     0,    57,    58,    12,     0,     0,
       0,     0,    61,     0,     0,     0,     0,     0,     0,    65,
       0,     0,     0,     0,     0,     0,     0,     0,     0,    45,
       0,     0,    66,    67,    47,    68,    69,     0,     0,    70,
       0,     0,    71,   127,     0,     0,     0,     0,    57,    58,
      12,     0,    72,    73,    74,    61,    13,    14,     0,     0,
       0,     0,   173,     0,     0,     0,     0,     0,     0,     0,
       0,     0,    45,     0,     0,    66,    67,   303,    68,   174,
       0,     0,    70,     0,     0,    71,   127,     0,     0,     0,
       0,    57,    58,    12,     0,    72,    73,    74,    61,    13,
      14,     0,     0,     0,     0,    65,     0,     0,     0,     0,
       0,     0,     0,     0,     0,     0,     0,     0,    66,    67,
      47,    68,    69,     0,     0,    70,     0,     0,    71,   127,
       0,     0,     0,     0,    57,    58,    12,     0,    72,    73,
      74,    61,    13,    14,     0,    47,     0,     0,   243,     0,
       0,     0,     0,    47,   127,     0,     0,     0,     0,    57,
      58,    12,   127,     0,     0,   129,    61,    57,    58,    12,
       0,   244,     0,   128,    61,     0,     0,   310,     0,    47,
       0,   243,     0,    74,     0,    13,    14,   301,   127,     0,
     129,     0,     0,    57,    58,    12,   130,     0,   129,     0,
      61,     0,     0,     0,   244,     0,    47,   431,    74,     0,
      13,    14,     0,     0,     0,   127,    74,     0,    13,    14,
      57,    58,    12,     0,   129,     0,     0,    61,     0,     0,
     432,     0,   303,     0,   243,     0,     0,     0,     0,     0,
     358,   127,    74,     0,    13,    14,    57,    58,    12,   359,
       0,   129,     0,    61,   360,   361,   362,   506,     0,     0,
     243,   363,     0,     0,     0,     0,     0,     0,   364,    74,
       0,    13,    14,     0,     0,     0,     0,   129,     0,     0,
       0,     0,     0,   244,     0,   365,     0,     0,     0,     0,
       0,   647,     0,     0,     0,    74,     0,    13,    14,     0,
       0,     0,     0,   367,   191,   192,    14,   193,     0,   195,
     196,   197,     0,     0,   199,   200,   201,   202,   203,   204,
     205,   206,   207,   208,   209,   210,   211,   212,     0,     0,
       0,     0,     0,     0,     0,     0,     0,   191,   192,     0,
     193,     0,   195,   196,   197,     0,   459,   199,   200,   201,
     202,   203,   204,   205,   206,   207,   208,   209,   210,   211,
     212,     0,     0,     0,     0,     0,     0,   191,   192,     0,
     193,     0,   195,   196,   197,     0,   456,   199,   200,   201,
     202,   203,   204,   205,   206,   207,   208,   209,   210,   211,
     212,   191,   192,     0,   193,     0,   195,   196,   197,     0,
     553,   199,   200,   201,   202,   203,   204,   205,   206,   207,
     208,   209,   210,   211,   212,   191,   192,     0,   193,     0,
     195,   196,   197,     0,   703,   199,   200,   201,   202,   203,
     204,   205,   206,   207,   208,   209,   210,   211,   212,   191,
     192,     0,   193,     0,   195,   196,   197,     0,   704,   199,
     200,   201,   202,   203,   204,   205,   206,   207,   208,   209,
     210,   211,   212,   191,   192,     0,     0,     0,   195,   196,
     197,     0,     0,   199,   200,   201,   202,   203,   204,   205,
     206,   207,   208,   209,   210,   211,   212,   191,   192,     0,
       0,     0,   195,   196,   197,     0,     0,   199,   200,   201,
     202,     0,   204,   205,   206,   207,   208,   209,   210,   211,
     212,   192,     0,     0,     0,   195,   196,   197,     0,     0,
     199,   200,   201,   202,     0,   204,   205,   206,   207,   208,
     209,   210,   211,   212
};

static const yytype_int16 yycheck[] =
{
       9,   216,   157,     6,    71,    77,   219,    20,   162,    47,
      19,   242,    47,   270,    71,   402,   158,   150,    31,   474,
      38,    30,   141,    26,   343,    28,   482,    46,    41,   515,
      49,   141,    41,   283,   284,    38,    55,   147,   269,   341,
      77,     3,   278,    46,    59,    11,    49,   157,     0,   336,
      53,   287,    55,    20,    24,   342,   414,     3,     3,   295,
      63,    64,    24,   299,   422,     3,   108,   109,    25,   111,
     112,    25,     7,   309,    77,    24,     3,    12,    24,    59,
      24,    24,     1,   128,   129,   130,    24,    59,    62,    50,
     634,    63,   105,    54,    74,   104,     5,    59,     3,    71,
      67,    63,     3,    73,    50,    21,   151,     5,   121,    75,
      59,    73,    74,    59,   159,   187,    35,   174,    66,    24,
     606,   166,    60,    24,    72,    63,   141,    73,    74,    73,
      74,   150,   490,    68,   230,    73,    74,   156,    24,   142,
     236,   237,   275,     5,   189,    72,    35,   150,    67,    50,
     187,    35,   696,   156,   698,   158,    65,   214,    63,   162,
      63,   463,   449,   413,   451,   415,    75,    65,    73,    74,
      59,    24,    73,    74,    24,   533,   407,    75,    67,    62,
     635,    62,    60,    67,   187,     1,   642,    68,   230,     3,
      62,   646,   647,     3,   236,   237,     7,    75,   684,   244,
      53,    12,    61,    65,    66,   462,    24,   222,   223,    59,
       3,    50,   215,    75,   227,    54,   261,   226,   221,    35,
      71,   239,    62,    73,    74,   438,   545,   109,   231,   111,
     112,    62,   445,   552,    93,   593,   239,    68,    67,   242,
      99,    59,   457,    59,   501,    66,   291,    35,    62,    63,
     253,    67,   610,   611,    68,    73,    74,    68,   303,     7,
      66,   306,   307,    59,    12,   303,   269,   270,   303,   365,
     337,    59,   426,     7,   393,   506,   409,   373,    12,    67,
     337,   639,   424,   393,   403,   300,    62,    59,    62,    35,
      59,    24,    68,   403,    68,   310,    74,    24,   301,    62,
      65,    66,    67,    68,    69,    70,    64,    72,    73,    24,
       9,    63,   548,    59,    62,    59,   358,    59,    17,    75,
      68,    67,    21,   365,    35,    35,    59,    67,    62,    35,
     717,   373,    31,    32,    68,    68,    60,    72,   341,   384,
      73,    74,   399,   356,    59,    59,   355,    72,    59,    59,
     669,     3,   609,    60,    35,     8,    67,    67,    73,    74,
      35,   399,    60,   366,   399,    64,   411,    24,    62,   441,
      24,    75,   603,   469,   377,    24,   391,   392,    60,   475,
     476,    24,   478,    62,    59,   452,   431,   432,    59,    62,
     409,   487,    67,   489,   397,   452,   399,    62,   417,    53,
      72,   404,   421,    60,   407,    59,   409,   620,   173,   174,
      59,   626,     3,    24,   417,   430,    73,    74,   421,    73,
      74,   424,    62,   426,    73,    74,    62,   469,    65,   444,
      62,    47,    67,   475,   476,    65,   478,    66,   441,    59,
      67,    67,   499,    71,     8,   487,    65,   489,    62,    60,
      62,    68,    62,    62,   467,    71,    24,   466,   515,   462,
     463,   499,    73,    74,   499,    60,    60,    60,   525,    68,
      35,   474,   475,    63,   477,    60,    60,   515,    60,   482,
     515,    60,   601,    60,    60,   488,   582,    75,   491,   492,
      68,   601,    60,    60,   590,    75,     3,    60,   501,    60,
      36,     8,    68,   506,   549,    73,    74,     3,    62,    60,
      17,    59,   128,   129,   130,    22,    23,    24,     8,    72,
      75,    62,    29,    60,    66,    60,   142,    17,    60,    60,
      24,    60,    22,    23,    24,   151,    62,    24,    60,    29,
     582,   637,   638,   159,   616,    60,    36,    59,   590,   606,
     166,    59,    59,    59,    68,   597,   569,   712,   174,   568,
      62,    72,    62,    53,    71,    59,    73,    74,   606,    59,
      24,   606,    59,   189,    49,    65,    68,    62,    59,    73,
      74,    71,    60,    73,    74,    75,    73,    74,    68,    60,
      60,    68,   634,   660,    14,   637,   638,    62,   214,    53,
     603,    60,   712,   660,    72,    59,   609,    60,    60,    60,
      68,    28,   616,   616,    68,   685,   538,   554,   631,    73,
      74,   630,   554,   301,   263,    49,   242,   684,   244,   263,
     174,   232,   635,   417,   637,   397,   409,    28,   641,   642,
     525,   358,   470,   646,   647,   261,   684,   263,   492,   684,
     358,    34,   641,   269,   696,   637,   698,   231,   488,    -1,
     673,    44,    -1,   672,    -1,    48,    49,    50,    51,    52,
      53,    54,    55,    56,    -1,   291,    -1,    -1,    -1,    -1,
      -1,   694,    47,    -1,   693,    -1,    -1,   303,    -1,    -1,
     306,   307,    -1,    -1,    -1,   708,    -1,    -1,   707,    -1,
      -1,    -1,   715,    -1,    -1,   714,    71,   720,    -1,    -1,
     719,    -1,   725,    -1,    -1,   724,   729,    -1,    -1,   728,
     733,   337,    -1,   732,   737,    -1,    -1,   736,   741,    -1,
      -1,   740,   745,    -1,    -1,   744,   749,    -1,    -1,   748,
     753,    -1,    -1,   752,   757,    -1,    -1,   756,   761,    -1,
     763,   760,    47,    -1,    -1,    -1,    -1,     8,    -1,    -1,
      11,    -1,    -1,   128,   129,   130,    17,    -1,   384,    -1,
      -1,    22,    23,    24,    -1,    -1,    -1,   142,    29,    -1,
      -1,   397,    -1,   399,    -1,    36,   151,    -1,    -1,    -1,
      -1,   407,     3,    -1,   159,   411,    -1,     8,    -1,    -1,
      11,   166,    53,    -1,    -1,    -1,    17,    34,    59,   174,
      -1,    22,    23,    24,    -1,   431,   432,    44,    29,    -1,
      71,    48,    73,    74,   189,    36,    53,    54,    55,    56,
      -1,    -1,    -1,   128,   129,   130,   452,    -1,    49,    50,
      -1,    52,    53,    -1,    -1,    56,    -1,   142,    59,   214,
      -1,    -1,     8,    -1,    -1,    -1,   151,    -1,    69,    70,
      71,    17,    73,    74,   159,    -1,    22,    23,    24,    -1,
      -1,   166,    -1,    29,    -1,    -1,    -1,   242,    -1,   244,
      36,    -1,    -1,   499,    -1,    -1,    -1,    -1,    -1,    -1,
     506,    -1,    -1,    -1,   189,    -1,   261,    53,   263,   515,
      -1,    -1,    -1,    59,   269,    -1,    -1,    -1,    -1,   525,
      -1,     3,    -1,    -1,    -1,    71,     8,    73,    74,    75,
      -1,    -1,    -1,    -1,    -1,    17,   291,    -1,    -1,    -1,
      22,    23,    24,   549,    26,    -1,    -1,    29,   303,    -1,
      -1,   306,   307,    -1,    36,    -1,    -1,   242,    -1,   244,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    49,    50,    -1,
      52,    53,    -1,    -1,    56,    -1,   261,    59,   263,    -1,
      -1,    -1,   337,    -1,   269,    -1,    -1,    69,    70,    71,
      -1,    73,    74,    -1,    -1,    -1,     8,   603,    -1,    -1,
     606,    -1,    -1,    -1,    -1,    17,   291,    -1,    -1,    -1,
      22,    23,    24,    -1,     3,    -1,    -1,    29,   303,     8,
      -1,   306,   307,    -1,    36,    -1,    -1,    -1,    17,   384,
      -1,    -1,    -1,    22,    23,    24,    -1,    -1,    -1,    -1,
      29,    53,   397,    -1,   399,    -1,    -1,    36,    -1,    -1,
      -1,    -1,   407,    65,   660,    -1,   411,    -1,    -1,    71,
      49,    50,    74,    52,    53,    -1,    -1,    56,    -1,    -1,
      59,    60,    -1,    -1,    -1,    -1,   431,   432,   684,     8,
      69,    70,    71,    -1,    73,    74,    -1,     8,    17,    -1,
      11,    -1,    -1,    22,    23,    24,    17,   452,    -1,   384,
      29,    22,    23,    24,    -1,    -1,    -1,    36,    29,    -1,
      -1,    -1,   397,    -1,   399,    36,    -1,    -1,    -1,    -1,
      -1,    -1,   407,    -1,    53,    -1,   411,    -1,    -1,    -1,
      59,    -1,    53,    -1,    -1,    -1,    -1,    59,    -1,    -1,
      -1,    -1,    71,    -1,   499,    74,   431,   432,    -1,    71,
      71,   506,    74,    74,    -1,    -1,    -1,    -1,    -1,    -1,
     515,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
     525,    -1,    -1,     0,     1,    -1,     3,    -1,    -1,     6,
      -1,     8,     9,    10,    -1,    -1,    13,    -1,    15,    16,
      17,    18,    19,    20,   549,    22,    23,    24,    -1,    -1,
      27,    28,    29,    30,    31,    32,    -1,    -1,    -1,    36,
      -1,    -1,    -1,    -1,   499,    -1,    -1,    -1,    -1,    -1,
      -1,   506,    49,    50,    -1,    52,    53,    -1,    -1,    56,
     515,    -1,    59,    -1,    -1,    62,    -1,    -1,    -1,    -1,
      -1,    -1,    69,    70,    71,    -1,    73,    74,   603,    -1,
      -1,   606,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,   549,    -1,    -1,    -1,   190,   191,
     192,   193,    -1,   195,   196,   197,    -1,   199,   200,   201,
     202,   203,   204,   205,   206,   207,   208,   209,   210,   211,
     212,    -1,   214,    -1,   216,    -1,   218,    -1,    -1,    -1,
     222,   223,   224,    -1,    -1,   660,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,   603,    -1,
      -1,   606,    -1,    -1,    -1,    -1,    -1,    -1,     1,   684,
       3,    -1,   254,     6,     7,     8,     9,    10,    -1,    12,
      13,    -1,    15,    16,    17,    18,    19,    20,    -1,    22,
      23,    24,    -1,    -1,    27,    28,    29,    30,    31,    32,
      -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    49,    50,   300,    52,
      53,    -1,    -1,    56,    -1,    -1,    59,    -1,   310,    62,
      -1,    -1,    -1,    -1,    67,    68,    69,    70,    71,   684,
      73,    74,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,   333,    -1,    -1,    -1,   337,    -1,    -1,    -1,    -1,
      -1,   343,    -1,     1,    -1,     3,    -1,    -1,     6,    -1,
       8,     9,    10,    -1,    -1,    13,    -1,    15,    16,    17,
      18,    19,    20,    -1,    22,    23,    24,    -1,    -1,    27,
      28,    29,    30,    31,    32,    -1,    -1,    -1,    36,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,   391,
     392,    49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,
      -1,    59,    -1,    -1,    62,    -1,    -1,    -1,    -1,    67,
      68,    69,    70,    71,    -1,    73,    74,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,     4,     5,    -1,    -1,   430,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,   444,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
     452,    -1,    -1,    33,    34,   457,    36,    37,    38,    39,
      40,    -1,    42,    43,    44,    45,    46,    47,    48,    49,
      50,    51,    52,    53,    54,    55,    56,    -1,     3,    -1,
      -1,    -1,    -1,     8,    -1,    65,    -1,    -1,    -1,    -1,
      -1,    -1,    17,     8,    -1,    75,    -1,    22,    23,    24,
     502,   503,    17,    -1,    29,    -1,    -1,    22,    23,    24,
      -1,    36,    -1,    -1,    29,    -1,    -1,    -1,    -1,    -1,
      -1,    36,    -1,   525,    49,    50,    -1,    52,    53,    -1,
      -1,    56,    -1,    -1,    59,    -1,    -1,    -1,    53,    -1,
      -1,    -1,    67,   545,    69,    70,    71,    -1,    73,    74,
     552,   553,   554,    -1,    -1,     3,    71,    -1,    73,    74,
       8,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    17,
      -1,    -1,    -1,    -1,    22,    23,    24,    -1,    -1,    -1,
      -1,    29,    -1,    31,    -1,    -1,    -1,    -1,    36,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    49,    50,    -1,    52,    53,    -1,    -1,    56,    -1,
      -1,    59,    -1,     3,    -1,    -1,    -1,    -1,     8,    -1,
      -1,    69,    70,    71,   626,    73,    74,    17,    -1,     8,
      -1,    -1,    22,    23,    24,    -1,    26,    -1,    17,    29,
      -1,    -1,    -1,    22,    23,    24,    36,    -1,    -1,    -1,
      29,    -1,    -1,    -1,    -1,   657,   658,    36,   660,    49,
      50,    -1,    52,    53,    -1,    -1,    56,   669,     3,    59,
      -1,    -1,    -1,     8,    53,    -1,    -1,    -1,    -1,    69,
      70,    71,    17,    73,    74,    -1,    -1,    22,    23,    24,
      -1,    26,    71,    -1,    29,    74,    -1,    -1,    -1,    -1,
      -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,     3,    -1,    -1,    49,    50,     8,    52,    53,    -1,
      -1,    56,    -1,    -1,    59,    17,    -1,    -1,    -1,    -1,
      22,    23,    24,    -1,    69,    70,    71,    29,    73,    74,
      -1,    -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    49,    50,    -1,
      52,    53,    -1,    -1,    56,    -1,     3,    59,    -1,    -1,
      -1,     8,    -1,    -1,    -1,    67,    -1,    69,    70,    71,
      17,    73,    74,    -1,    -1,    22,    23,    24,    -1,    -1,
      -1,    -1,    29,    -1,    -1,    -1,    -1,    -1,    -1,    36,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,     3,
      -1,    -1,    49,    50,     8,    52,    53,    -1,    -1,    56,
      -1,    -1,    59,    17,    -1,    -1,    -1,    -1,    22,    23,
      24,    -1,    69,    70,    71,    29,    73,    74,    -1,    -1,
      -1,    -1,    36,    -1,    -1,    -1,    -1,    -1,    -1,    -1,
      -1,    -1,     3,    -1,    -1,    49,    50,     8,    52,    53,
      -1,    -1,    56,    -1,    -1,    59,    17,    -1,    -1,    -1,
      -1,    22,    23,    24,    -1,    69,    70,    71,    29,    73,
      74,    -1,    -1,    -1,    -1,    36,    -1,    -1,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    -1,    49,    50,
       8,    52,    53,    -1,    -1,    56,    -1,    -1,    59,    17,
      -1,    -1,    -1,    -1,    22,    23,    24,    -1,    69,    70,
      71,    29,    73,    74,    -1,     8,    -1,    -1,    36,    -1,
      -1,    -1,    -1,     8,    17,    -1,    -1,    -1,    -1,    22,
      23,    24,    17,    -1,    -1,    53,    29,    22,    23,    24,
      -1,    59,    -1,    36,    29,    -1,    -1,    65,    -1,     8,
      -1,    36,    -1,    71,    -1,    73,    74,    75,    17,    -1,
      53,    -1,    -1,    22,    23,    24,    59,    -1,    53,    -1,
      29,    -1,    -1,    -1,    59,    -1,     8,    36,    71,    -1,
      73,    74,    -1,    -1,    -1,    17,    71,    -1,    73,    74,
      22,    23,    24,    -1,    53,    -1,    -1,    29,    -1,    -1,
      59,    -1,     8,    -1,    36,    -1,    -1,    -1,    -1,    -1,
       8,    17,    71,    -1,    73,    74,    22,    23,    24,    17,
      -1,    53,    -1,    29,    22,    23,    24,    59,    -1,    -1,
      36,    29,    -1,    -1,    -1,    -1,    -1,    -1,    36,    71,
      -1,    73,    74,    -1,    -1,    -1,    -1,    53,    -1,    -1,
      -1,    -1,    -1,    59,    -1,    53,    -1,    -1,    -1,    -1,
      -1,    59,    -1,    -1,    -1,    71,    -1,    73,    74,    -1,
      -1,    -1,    -1,    71,    33,    34,    74,    36,    -1,    38,
      39,    40,    -1,    -1,    43,    44,    45,    46,    47,    48,
      49,    50,    51,    52,    53,    54,    55,    56,    -1,    -1,
      -1,    -1,    -1,    -1,    -1,    -1,    -1,    33,    34,    -1,
      36,    -1,    38,    39,    40,    -1,    75,    43,    44,    45,
      46,    47,    48,    49,    50,    51,    52,    53,    54,    55,
      56,    -1,    -1,    -1,    -1,    -1,    -1,    33,    34,    -1,
      36,    -1,    38,    39,    40,    -1,    72,    43,    44,    45,
      46,    47,    48,    49,    50,    51,    52,    53,    54,    55,
      56,    33,    34,    -1,    36,    -1,    38,    39,    40,    -1,
      66,    43,    44,    45,    46,    47,    48,    49,    50,    51,
      52,    53,    54,    55,    56,    33,    34,    -1,    36,    -1,
      38,    39,    40,    -1,    66,    43,    44,    45,    46,    47,
      48,    49,    50,    51,    52,    53,    54,    55,    56,    33,
      34,    -1,    36,    -1,    38,    39,    40,    -1,    66,    43,
      44,    45,    46,    47,    48,    49,    50,    51,    52,    53,
      54,    55,    56,    33,    34,    -1,    -1,    -1,    38,    39,
      40,    -1,    -1,    43,    44,    45,    46,    47,    48,    49,
      50,    51,    52,    53,    54,    55,    56,    33,    34,    -1,
      -1,    -1,    38,    39,    40,    -1,    -1,    43,    44,    45,
      46,    -1,    48,    49,    50,    51,    52,    53,    54,    55,
      56,    34,    -1,    -1,    -1,    38,    39,    40,    -1,    -1,
      43,    44,    45,    46,    -1,    48,    49,    50,    51,    52,
      53,    54,    55,    56
};

/* YYSTOS[STATE-NUM] -- The (internal number of the) accessing
   symbol of state STATE-NUM.  */
static const yytype_uint16 yystos[] =
{
       0,    77,    79,    80,    81,     0,    25,    78,    82,    83,
      25,   135,    24,    73,    74,   190,   191,   130,    84,    85,
     135,    24,   137,   138,     3,    62,    21,   131,   215,    86,
      87,   135,   137,    24,   136,   263,    63,     3,    59,    63,
     132,   134,   190,    62,     1,     3,     6,     8,     9,    10,
      13,    15,    16,    17,    18,    19,    20,    22,    23,    27,
      28,    29,    30,    31,    32,    36,    49,    50,    52,    53,
      56,    59,    69,    70,    71,   139,   140,   141,   147,   159,
     162,   170,   173,   175,   176,   177,   178,   183,   187,   190,
     192,   193,   198,   199,   202,   205,   206,   207,   210,   213,
     214,   230,   235,    88,    89,   135,   137,    62,     9,    17,
      21,    31,    32,    64,   248,    24,    73,    60,   132,   133,
       3,   135,   137,     3,   187,   189,   190,    17,    36,    53,
      59,   190,   192,   197,   201,   202,   203,   210,   189,   177,
     183,   160,    59,   190,   208,   177,   187,   163,    35,    67,
     186,    71,   175,   235,   242,   174,   186,   171,    59,   145,
     146,   190,    59,   142,   188,   190,   234,   176,   176,   176,
     176,   176,   176,    36,    53,   175,   184,   196,   202,   204,
     210,   176,   176,    11,   175,   241,    62,    59,   143,   234,
       4,    33,    34,    36,    37,    38,    39,    40,    42,    43,
      44,    45,    46,    47,    48,    49,    50,    51,    52,    53,
      54,    55,    56,    67,    59,    63,    71,    66,    59,   186,
       1,   186,     5,    65,    75,    90,    91,   135,   137,   191,
     249,    59,   209,   249,    24,   249,   250,   249,    64,    62,
     239,   137,    59,    36,    59,   195,   201,   202,   203,   204,
     210,   195,   195,    63,    26,   147,   156,   157,   158,   235,
     243,    11,   185,   190,   194,   195,   226,   227,   228,    59,
      67,   211,   161,   243,    24,    59,    68,   187,   220,   222,
     224,   195,    35,    53,    59,    68,   187,   219,   221,   222,
     223,   233,   161,    60,   146,   218,   195,    60,   142,   216,
      65,    75,   195,     8,   196,    60,    72,    72,    60,   143,
      65,   195,   175,   175,   175,   175,   175,   175,   175,   175,
     175,   175,   175,   175,   175,   175,   175,   175,   175,   175,
     175,   175,   175,   179,    60,   184,   236,    59,   190,   175,
     241,   231,   175,   179,     1,    67,   140,   149,   229,   230,
     232,   235,   235,   175,    92,    93,   135,   137,     8,    17,
      22,    23,    24,    29,    36,    53,    65,    71,   191,   251,
     253,   254,   255,   190,   256,   264,   211,    59,     3,   251,
     251,   132,    60,   228,     8,   195,    60,   190,   175,    35,
     154,     5,    65,    62,   195,   185,   194,    75,   240,    60,
     228,   232,   164,    62,    63,    24,   222,    59,   225,    62,
     239,    72,   153,    59,   223,    53,   223,    62,   239,     3,
     247,    75,   195,   172,    62,   239,    62,   239,   235,   188,
      65,    36,    59,   195,   201,   202,   203,   210,    67,   195,
     195,    62,   239,   235,    65,    67,   175,   180,   181,   237,
     238,    11,    75,   240,    31,   184,    72,    66,   229,    75,
     240,   238,   150,    62,    68,    94,    95,   135,   137,    36,
      59,   252,   253,   255,    59,    67,    71,    67,     8,   251,
       3,    50,    59,   190,   261,   262,     3,    72,    65,    11,
     251,    60,    75,    62,   244,   264,    62,    62,    62,    60,
      60,   155,    26,    26,   243,   226,    59,   190,   200,   201,
     202,   203,   204,   210,   212,    60,    68,   154,   243,   190,
      60,   228,   224,    68,   195,     7,    12,    68,   148,   151,
     223,   247,   223,    60,   221,    68,   187,   247,    35,   146,
      60,   142,    60,   235,   195,   179,   143,   144,   217,   234,
      60,   235,   179,    66,    75,   240,    68,   240,   184,    60,
      60,    60,   241,    60,    68,   232,   229,    96,    97,   135,
     137,   251,   254,   244,    24,   190,   191,   246,   251,   258,
     266,   251,   190,   245,   257,   265,   251,     3,   261,    62,
      72,   251,   262,   251,   247,   190,   256,    60,   232,   175,
     175,    62,   228,    59,   212,   165,    60,   236,    66,   152,
      60,    60,   247,   153,    60,   238,    62,   239,   195,   238,
      67,   175,   182,   180,   181,    60,    66,    72,    68,    98,
      99,   135,   137,    60,    60,    59,    68,    62,    72,   251,
      68,    62,    49,   251,    62,   247,    59,    59,   251,   259,
     260,    68,   243,    60,   228,   168,   212,     5,    65,    66,
      75,   232,   247,   247,    68,    68,   144,    60,    68,   179,
     241,   100,   101,   135,   137,   259,   244,   258,   251,   247,
     257,   261,   244,   244,    60,    14,   166,   169,   175,   175,
     238,    72,   102,   103,   135,   137,    60,    60,    60,    60,
     212,    20,   149,    66,    66,    68,   104,   105,   135,   137,
     259,   259,   167,   106,   107,   135,   137,   161,   108,   109,
     135,   137,   154,   110,   111,   135,   137,   112,   113,   135,
     137,   114,   115,   135,   137,   116,   117,   135,   137,   118,
     119,   135,   137,   120,   121,   135,   137,   122,   123,   135,
     137,   124,   125,   135,   137,   126,   127,   135,   137,   128,
     129,   135,   137,   135,   137,   137
};

#define yyerrok		(yyerrstatus = 0)
#define yyclearin	(yychar = YYEMPTY)
#define YYEMPTY		(-2)
#define YYEOF		0

#define YYACCEPT	goto yyacceptlab
#define YYABORT		goto yyabortlab
#define YYERROR		goto yyerrorlab


/* Like YYERROR except do call yyerror.  This remains here temporarily
   to ease the transition to the new meaning of YYERROR, for GCC.
   Once GCC version 2 has supplanted version 1, this can go.  */

#define YYFAIL		goto yyerrlab

#define YYRECOVERING()  (!!yyerrstatus)

#define YYBACKUP(Token, Value)					\
do								\
  if (yychar == YYEMPTY && yylen == 1)				\
    {								\
      yychar = (Token);						\
      yylval = (Value);						\
      yytoken = YYTRANSLATE (yychar);				\
      YYPOPSTACK (1);						\
      goto yybackup;						\
    }								\
  else								\
    {								\
      yyerror (YY_("syntax error: cannot back up")); \
      YYERROR;							\
    }								\
while (YYID (0))


#define YYTERROR	1
#define YYERRCODE	256


/* YYLLOC_DEFAULT -- Set CURRENT to span from RHS[1] to RHS[N].
   If N is 0, then set CURRENT to the empty location which ends
   the previous symbol: RHS[0] (always defined).  */

#define YYRHSLOC(Rhs, K) ((Rhs)[K])
#ifndef YYLLOC_DEFAULT
# define YYLLOC_DEFAULT(Current, Rhs, N)				\
    do									\
      if (YYID (N))                                                    \
	{								\
	  (Current).first_line   = YYRHSLOC (Rhs, 1).first_line;	\
	  (Current).first_column = YYRHSLOC (Rhs, 1).first_column;	\
	  (Current).last_line    = YYRHSLOC (Rhs, N).last_line;		\
	  (Current).last_column  = YYRHSLOC (Rhs, N).last_column;	\
	}								\
      else								\
	{								\
	  (Current).first_line   = (Current).last_line   =		\
	    YYRHSLOC (Rhs, 0).last_line;				\
	  (Current).first_column = (Current).last_column =		\
	    YYRHSLOC (Rhs, 0).last_column;				\
	}								\
    while (YYID (0))
#endif


/* YY_LOCATION_PRINT -- Print the location on the stream.
   This macro was not mandated originally: define only if we know
   we won't break user code: when these are the locations we know.  */

#ifndef YY_LOCATION_PRINT
# if defined YYLTYPE_IS_TRIVIAL && YYLTYPE_IS_TRIVIAL
#  define YY_LOCATION_PRINT(File, Loc)			\
     fprintf (File, "%d.%d-%d.%d",			\
	      (Loc).first_line, (Loc).first_column,	\
	      (Loc).last_line,  (Loc).last_column)
# else
#  define YY_LOCATION_PRINT(File, Loc) ((void) 0)
# endif
#endif


/* YYLEX -- calling `yylex' with the right arguments.  */

#ifdef YYLEX_PARAM
# define YYLEX yylex (YYLEX_PARAM)
#else
# define YYLEX yylex ()
#endif

/* Enable debugging if requested.  */
#if YYDEBUG

# ifndef YYFPRINTF
#  include <stdio.h> /* INFRINGES ON USER NAME SPACE */
#  define YYFPRINTF fprintf
# endif

# define YYDPRINTF(Args)			\
do {						\
  if (yydebug)					\
    YYFPRINTF Args;				\
} while (YYID (0))

# define YY_SYMBOL_PRINT(Title, Type, Value, Location)			  \
do {									  \
  if (yydebug)								  \
    {									  \
      YYFPRINTF (stderr, "%s ", Title);					  \
      yy_symbol_print (stderr,						  \
		  Type, Value); \
      YYFPRINTF (stderr, "\n");						  \
    }									  \
} while (YYID (0))


/*--------------------------------.
| Print this symbol on YYOUTPUT.  |
`--------------------------------*/

/*ARGSUSED*/
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_symbol_value_print (FILE *yyoutput, int yytype, YYSTYPE const * const yyvaluep)
#else
static void
yy_symbol_value_print (yyoutput, yytype, yyvaluep)
    FILE *yyoutput;
    int yytype;
    YYSTYPE const * const yyvaluep;
#endif
{
  if (!yyvaluep)
    return;
# ifdef YYPRINT
  if (yytype < YYNTOKENS)
    YYPRINT (yyoutput, yytoknum[yytype], *yyvaluep);
# else
  YYUSE (yyoutput);
# endif
  switch (yytype)
    {
      default:
	break;
    }
}


/*--------------------------------.
| Print this symbol on YYOUTPUT.  |
`--------------------------------*/

#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_symbol_print (FILE *yyoutput, int yytype, YYSTYPE const * const yyvaluep)
#else
static void
yy_symbol_print (yyoutput, yytype, yyvaluep)
    FILE *yyoutput;
    int yytype;
    YYSTYPE const * const yyvaluep;
#endif
{
  if (yytype < YYNTOKENS)
    YYFPRINTF (yyoutput, "token %s (", yytname[yytype]);
  else
    YYFPRINTF (yyoutput, "nterm %s (", yytname[yytype]);

  yy_symbol_value_print (yyoutput, yytype, yyvaluep);
  YYFPRINTF (yyoutput, ")");
}

/*------------------------------------------------------------------.
| yy_stack_print -- Print the state stack from its BOTTOM up to its |
| TOP (included).                                                   |
`------------------------------------------------------------------*/

#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_stack_print (yytype_int16 *bottom, yytype_int16 *top)
#else
static void
yy_stack_print (bottom, top)
    yytype_int16 *bottom;
    yytype_int16 *top;
#endif
{
  YYFPRINTF (stderr, "Stack now");
  for (; bottom <= top; ++bottom)
    YYFPRINTF (stderr, " %d", *bottom);
  YYFPRINTF (stderr, "\n");
}

# define YY_STACK_PRINT(Bottom, Top)				\
do {								\
  if (yydebug)							\
    yy_stack_print ((Bottom), (Top));				\
} while (YYID (0))


/*------------------------------------------------.
| Report that the YYRULE is going to be reduced.  |
`------------------------------------------------*/

#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yy_reduce_print (YYSTYPE *yyvsp, int yyrule)
#else
static void
yy_reduce_print (yyvsp, yyrule)
    YYSTYPE *yyvsp;
    int yyrule;
#endif
{
  int yynrhs = yyr2[yyrule];
  int yyi;
  unsigned long int yylno = yyrline[yyrule];
  YYFPRINTF (stderr, "Reducing stack by rule %d (line %lu):\n",
	     yyrule - 1, yylno);
  /* The symbols being reduced.  */
  for (yyi = 0; yyi < yynrhs; yyi++)
    {
      fprintf (stderr, "   $%d = ", yyi + 1);
      yy_symbol_print (stderr, yyrhs[yyprhs[yyrule] + yyi],
		       &(yyvsp[(yyi + 1) - (yynrhs)])
		       		       );
      fprintf (stderr, "\n");
    }
}

# define YY_REDUCE_PRINT(Rule)		\
do {					\
  if (yydebug)				\
    yy_reduce_print (yyvsp, Rule); \
} while (YYID (0))

/* Nonzero means print parse trace.  It is left uninitialized so that
   multiple parsers can coexist.  */
int yydebug;
#else /* !YYDEBUG */
# define YYDPRINTF(Args)
# define YY_SYMBOL_PRINT(Title, Type, Value, Location)
# define YY_STACK_PRINT(Bottom, Top)
# define YY_REDUCE_PRINT(Rule)
#endif /* !YYDEBUG */


/* YYINITDEPTH -- initial size of the parser's stacks.  */
#ifndef	YYINITDEPTH
# define YYINITDEPTH 200
#endif

/* YYMAXDEPTH -- maximum size the stacks can grow to (effective only
   if the built-in stack extension method is used).

   Do not make this value too large; the results are undefined if
   YYSTACK_ALLOC_MAXIMUM < YYSTACK_BYTES (YYMAXDEPTH)
   evaluated with infinite-precision integer arithmetic.  */

#ifndef YYMAXDEPTH
# define YYMAXDEPTH 10000
#endif



#if YYERROR_VERBOSE

# ifndef yystrlen
#  if defined __GLIBC__ && defined _STRING_H
#   define yystrlen strlen
#  else
/* Return the length of YYSTR.  */
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static YYSIZE_T
yystrlen (const char *yystr)
#else
static YYSIZE_T
yystrlen (yystr)
    const char *yystr;
#endif
{
  YYSIZE_T yylen;
  for (yylen = 0; yystr[yylen]; yylen++)
    continue;
  return yylen;
}
#  endif
# endif

# ifndef yystpcpy
#  if defined __GLIBC__ && defined _STRING_H && defined _GNU_SOURCE
#   define yystpcpy stpcpy
#  else
/* Copy YYSRC to YYDEST, returning the address of the terminating '\0' in
   YYDEST.  */
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static char *
yystpcpy (char *yydest, const char *yysrc)
#else
static char *
yystpcpy (yydest, yysrc)
    char *yydest;
    const char *yysrc;
#endif
{
  char *yyd = yydest;
  const char *yys = yysrc;

  while ((*yyd++ = *yys++) != '\0')
    continue;

  return yyd - 1;
}
#  endif
# endif

# ifndef yytnamerr
/* Copy to YYRES the contents of YYSTR after stripping away unnecessary
   quotes and backslashes, so that it's suitable for yyerror.  The
   heuristic is that double-quoting is unnecessary unless the string
   contains an apostrophe, a comma, or backslash (other than
   backslash-backslash).  YYSTR is taken from yytname.  If YYRES is
   null, do not copy; instead, return the length of what the result
   would have been.  */
static YYSIZE_T
yytnamerr (char *yyres, const char *yystr)
{
  if (*yystr == '"')
    {
      YYSIZE_T yyn = 0;
      char const *yyp = yystr;

      for (;;)
	switch (*++yyp)
	  {
	  case '\'':
	  case ',':
	    goto do_not_strip_quotes;

	  case '\\':
	    if (*++yyp != '\\')
	      goto do_not_strip_quotes;
	    /* Fall through.  */
	  default:
	    if (yyres)
	      yyres[yyn] = *yyp;
	    yyn++;
	    break;

	  case '"':
	    if (yyres)
	      yyres[yyn] = '\0';
	    return yyn;
	  }
    do_not_strip_quotes: ;
    }

  if (! yyres)
    return yystrlen (yystr);

  return yystpcpy (yyres, yystr) - yyres;
}
# endif

/* Copy into YYRESULT an error message about the unexpected token
   YYCHAR while in state YYSTATE.  Return the number of bytes copied,
   including the terminating null byte.  If YYRESULT is null, do not
   copy anything; just return the number of bytes that would be
   copied.  As a special case, return 0 if an ordinary "syntax error"
   message will do.  Return YYSIZE_MAXIMUM if overflow occurs during
   size calculation.  */
static YYSIZE_T
yysyntax_error (char *yyresult, int yystate, int yychar)
{
  int yyn = yypact[yystate];

  if (! (YYPACT_NINF < yyn && yyn <= YYLAST))
    return 0;
  else
    {
      int yytype = YYTRANSLATE (yychar);
      YYSIZE_T yysize0 = yytnamerr (0, yytname[yytype]);
      YYSIZE_T yysize = yysize0;
      YYSIZE_T yysize1;
      int yysize_overflow = 0;
      enum { YYERROR_VERBOSE_ARGS_MAXIMUM = 5 };
      char const *yyarg[YYERROR_VERBOSE_ARGS_MAXIMUM];
      int yyx;

# if 0
      /* This is so xgettext sees the translatable formats that are
	 constructed on the fly.  */
      YY_("syntax error, unexpected %s");
      YY_("syntax error, unexpected %s, expecting %s");
      YY_("syntax error, unexpected %s, expecting %s or %s");
      YY_("syntax error, unexpected %s, expecting %s or %s or %s");
      YY_("syntax error, unexpected %s, expecting %s or %s or %s or %s");
# endif
      char *yyfmt;
      char const *yyf;
      static char const yyunexpected[] = "syntax error, unexpected %s";
      static char const yyexpecting[] = ", expecting %s";
      static char const yyor[] = " or %s";
      char yyformat[sizeof yyunexpected
		    + sizeof yyexpecting - 1
		    + ((YYERROR_VERBOSE_ARGS_MAXIMUM - 2)
		       * (sizeof yyor - 1))];
      char const *yyprefix = yyexpecting;

      /* Start YYX at -YYN if negative to avoid negative indexes in
	 YYCHECK.  */
      int yyxbegin = yyn < 0 ? -yyn : 0;

      /* Stay within bounds of both yycheck and yytname.  */
      int yychecklim = YYLAST - yyn + 1;
      int yyxend = yychecklim < YYNTOKENS ? yychecklim : YYNTOKENS;
      int yycount = 1;

      yyarg[0] = yytname[yytype];
      yyfmt = yystpcpy (yyformat, yyunexpected);

      for (yyx = yyxbegin; yyx < yyxend; ++yyx)
	if (yycheck[yyx + yyn] == yyx && yyx != YYTERROR)
	  {
	    if (yycount == YYERROR_VERBOSE_ARGS_MAXIMUM)
	      {
		yycount = 1;
		yysize = yysize0;
		yyformat[sizeof yyunexpected - 1] = '\0';
		break;
	      }
	    yyarg[yycount++] = yytname[yyx];
	    yysize1 = yysize + yytnamerr (0, yytname[yyx]);
	    yysize_overflow |= (yysize1 < yysize);
	    yysize = yysize1;
	    yyfmt = yystpcpy (yyfmt, yyprefix);
	    yyprefix = yyor;
	  }

      yyf = YY_(yyformat);
      yysize1 = yysize + yystrlen (yyf);
      yysize_overflow |= (yysize1 < yysize);
      yysize = yysize1;

      if (yysize_overflow)
	return YYSIZE_MAXIMUM;

      if (yyresult)
	{
	  /* Avoid sprintf, as that infringes on the user's name space.
	     Don't have undefined behavior even if the translation
	     produced a string with the wrong number of "%s"s.  */
	  char *yyp = yyresult;
	  int yyi = 0;
	  while ((*yyp = *yyf) != '\0')
	    {
	      if (*yyp == '%' && yyf[1] == 's' && yyi < yycount)
		{
		  yyp += yytnamerr (yyp, yyarg[yyi++]);
		  yyf += 2;
		}
	      else
		{
		  yyp++;
		  yyf++;
		}
	    }
	}
      return yysize;
    }
}
#endif /* YYERROR_VERBOSE */


/*-----------------------------------------------.
| Release the memory associated to this symbol.  |
`-----------------------------------------------*/

/*ARGSUSED*/
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
static void
yydestruct (const char *yymsg, int yytype, YYSTYPE *yyvaluep)
#else
static void
yydestruct (yymsg, yytype, yyvaluep)
    const char *yymsg;
    int yytype;
    YYSTYPE *yyvaluep;
#endif
{
  YYUSE (yyvaluep);

  if (!yymsg)
    yymsg = "Deleting";
  YY_SYMBOL_PRINT (yymsg, yytype, yyvaluep, yylocationp);

  switch (yytype)
    {

      default:
	break;
    }
}


/* Prevent warnings from -Wmissing-prototypes.  */

#ifdef YYPARSE_PARAM
#if defined __STDC__ || defined __cplusplus
int yyparse (void *YYPARSE_PARAM);
#else
int yyparse ();
#endif
#else /* ! YYPARSE_PARAM */
#if defined __STDC__ || defined __cplusplus
int yyparse (void);
#else
int yyparse ();
#endif
#endif /* ! YYPARSE_PARAM */



/* The look-ahead symbol.  */
int yychar, yystate;

/* The semantic value of the look-ahead symbol.  */
YYSTYPE yylval;

/* Number of syntax errors so far.  */
int yynerrs;



/*----------.
| yyparse.  |
`----------*/

#ifdef YYPARSE_PARAM
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
int
yyparse (void *YYPARSE_PARAM)
#else
int
yyparse (YYPARSE_PARAM)
    void *YYPARSE_PARAM;
#endif
#else /* ! YYPARSE_PARAM */
#if (defined __STDC__ || defined __C99__FUNC__ \
     || defined __cplusplus || defined _MSC_VER)
int
yyparse (void)
#else
int
yyparse ()

#endif
#endif
{
  
  int yyn;
  int yyresult;
  /* Number of tokens to shift before error messages enabled.  */
  int yyerrstatus;
  /* Look-ahead token as an internal (translated) token number.  */
  int yytoken = 0;
#if YYERROR_VERBOSE
  /* Buffer for error messages, and its allocated size.  */
  char yymsgbuf[128];
  char *yymsg = yymsgbuf;
  YYSIZE_T yymsg_alloc = sizeof yymsgbuf;
#endif

  /* Three stacks and their tools:
     `yyss': related to states,
     `yyvs': related to semantic values,
     `yyls': related to locations.

     Refer to the stacks thru separate pointers, to allow yyoverflow
     to reallocate them elsewhere.  */

  /* The state stack.  */
  yytype_int16 yyssa[YYINITDEPTH];
  yytype_int16 *yyss = yyssa;
  yytype_int16 *yyssp;

  /* The semantic value stack.  */
  YYSTYPE yyvsa[YYINITDEPTH];
  YYSTYPE *yyvs = yyvsa;
  YYSTYPE *yyvsp;



#define YYPOPSTACK(N)   (yyvsp -= (N), yyssp -= (N))

  YYSIZE_T yystacksize = YYINITDEPTH;

  /* The variables used to return semantic value and location from the
     action routines.  */
  YYSTYPE yyval;


  /* The number of symbols on the RHS of the reduced rule.
     Keep to zero when no symbol should be popped.  */
  int yylen = 0;

  YYDPRINTF ((stderr, "Starting parse\n"));

  yystate = 0;
  yyerrstatus = 0;
  yynerrs = 0;
  yychar = YYEMPTY;		/* Cause a token to be read.  */

  /* Initialize stack pointers.
     Waste one element of value and location stack
     so that they stay on the same level as the state stack.
     The wasted elements are never initialized.  */

  yyssp = yyss;
  yyvsp = yyvs;

  goto yysetstate;

/*------------------------------------------------------------.
| yynewstate -- Push a new state, which is found in yystate.  |
`------------------------------------------------------------*/
 yynewstate:
  /* In all cases, when you get here, the value and location stacks
     have just been pushed.  So pushing a state here evens the stacks.  */
  yyssp++;

 yysetstate:
  *yyssp = yystate;

  if (yyss + yystacksize - 1 <= yyssp)
    {
      /* Get the current used size of the three stacks, in elements.  */
      YYSIZE_T yysize = yyssp - yyss + 1;

#ifdef yyoverflow
      {
	/* Give user a chance to reallocate the stack.  Use copies of
	   these so that the &'s don't force the real ones into
	   memory.  */
	YYSTYPE *yyvs1 = yyvs;
	yytype_int16 *yyss1 = yyss;


	/* Each stack pointer address is followed by the size of the
	   data in use in that stack, in bytes.  This used to be a
	   conditional around just the two extra args, but that might
	   be undefined if yyoverflow is a macro.  */
	yyoverflow (YY_("memory exhausted"),
		    &yyss1, yysize * sizeof (*yyssp),
		    &yyvs1, yysize * sizeof (*yyvsp),

		    &yystacksize);

	yyss = yyss1;
	yyvs = yyvs1;
      }
#else /* no yyoverflow */
# ifndef YYSTACK_RELOCATE
      goto yyexhaustedlab;
# else
      /* Extend the stack our own way.  */
      if (YYMAXDEPTH <= yystacksize)
	goto yyexhaustedlab;
      yystacksize *= 2;
      if (YYMAXDEPTH < yystacksize)
	yystacksize = YYMAXDEPTH;

      {
	yytype_int16 *yyss1 = yyss;
	union yyalloc *yyptr =
	  (union yyalloc *) YYSTACK_ALLOC (YYSTACK_BYTES (yystacksize));
	if (! yyptr)
	  goto yyexhaustedlab;
	YYSTACK_RELOCATE (yyss);
	YYSTACK_RELOCATE (yyvs);

#  undef YYSTACK_RELOCATE
	if (yyss1 != yyssa)
	  YYSTACK_FREE (yyss1);
      }
# endif
#endif /* no yyoverflow */

      yyssp = yyss + yysize - 1;
      yyvsp = yyvs + yysize - 1;


      YYDPRINTF ((stderr, "Stack size increased to %lu\n",
		  (unsigned long int) yystacksize));

      if (yyss + yystacksize - 1 <= yyssp)
	YYABORT;
    }

  YYDPRINTF ((stderr, "Entering state %d\n", yystate));

  goto yybackup;

/*-----------.
| yybackup.  |
`-----------*/
yybackup:

  /* Do appropriate processing given the current state.  Read a
     look-ahead token if we need one and don't already have one.  */

  /* First try to decide what to do without reference to look-ahead token.  */
  yyn = yypact[yystate];
  if (yyn == YYPACT_NINF)
    goto yydefault;

  /* Not known => get a look-ahead token if don't already have one.  */

  /* YYCHAR is either YYEMPTY or YYEOF or a valid look-ahead symbol.  */
  if (yychar == YYEMPTY)
    {
      YYDPRINTF ((stderr, "Reading a token: "));
      yychar = YYLEX;
    }

  if (yychar <= YYEOF)
    {
      yychar = yytoken = YYEOF;
      YYDPRINTF ((stderr, "Now at end of input.\n"));
    }
  else
    {
      yytoken = YYTRANSLATE (yychar);
      YY_SYMBOL_PRINT ("Next token is", yytoken, &yylval, &yylloc);
    }

  /* If the proper action on seeing token YYTOKEN is to reduce or to
     detect an error, take that action.  */
  yyn += yytoken;
  if (yyn < 0 || YYLAST < yyn || yycheck[yyn] != yytoken)
    goto yydefault;
  yyn = yytable[yyn];
  if (yyn <= 0)
    {
      if (yyn == 0 || yyn == YYTABLE_NINF)
	goto yyerrlab;
      yyn = -yyn;
      goto yyreduce;
    }

  if (yyn == YYFINAL)
    YYACCEPT;

  /* Count tokens shifted since error; after three, turn off error
     status.  */
  if (yyerrstatus)
    yyerrstatus--;

  /* Shift the look-ahead token.  */
  YY_SYMBOL_PRINT ("Shifting", yytoken, &yylval, &yylloc);

  /* Discard the shifted token unless it is eof.  */
  if (yychar != YYEOF)
    yychar = YYEMPTY;

  yystate = yyn;
  *++yyvsp = yylval;

  goto yynewstate;


/*-----------------------------------------------------------.
| yydefault -- do the default action for the current state.  |
`-----------------------------------------------------------*/
yydefault:
  yyn = yydefact[yystate];
  if (yyn == 0)
    goto yyerrlab;
  goto yyreduce;


/*-----------------------------.
| yyreduce -- Do a reduction.  |
`-----------------------------*/
yyreduce:
  /* yyn is the number of a rule to reduce with.  */
  yylen = yyr2[yyn];

  /* If YYLEN is nonzero, implement the default value of the action:
     `$$ = $1'.

     Otherwise, the following line sets YYVAL to garbage.
     This behavior is undocumented and Bison
     users should not rely upon it.  Assigning to YYVAL
     unconditionally makes the parser a bit smaller, and it avoids a
     GCC warning that YYVAL may be used uninitialized.  */
  yyval = yyvsp[1-yylen];


  YY_REDUCE_PRINT (yyn);
  switch (yyn)
    {
        case 2:
#line 128 "go.y"
    {
		xtop = concat(xtop, (yyvsp[(4) - (4)].list));
	}
    break;

  case 3:
#line 134 "go.y"
    {
		prevlineno = lineno;
		yyerror("package statement must be first");
		errorexit();
	}
    break;

  case 4:
#line 140 "go.y"
    {
		mkpackage((yyvsp[(2) - (3)].sym)->name);
	}
    break;

  case 5:
#line 175 "go.y"
    {
	}
    break;

  case 6:
#line 179 "go.y"
    {
		importpkg = corepkg;
		if(debug['A']) {
			cannedimports("core.builtin", "package core\n\n$$\n\n");
		} else {
			cannedimports("core.builtin", coreimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 7:
#line 191 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 8:
#line 196 "go.y"
    {
		importpkg = lockpkg;
		if(debug['A']) {
			cannedimports("lock.builtin", "package lock\n\n$$\n\n");
		} else {
			cannedimports("lock.builtin", lockimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 9:
#line 208 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 10:
#line 213 "go.y"
    {
		importpkg = schedpkg;
		if(debug['A']) {
			cannedimports("sched.builtin", "package sched\n\n$$\n\n");
		} else {
			cannedimports("sched.builtin", schedimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 11:
#line 225 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 12:
#line 230 "go.y"
    {
		importpkg = sempkg;
		if(debug['A']) {
			cannedimports("sem.builtin", "package sem\n\n$$\n\n");
		} else {
			cannedimports("sem.builtin", semimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 13:
#line 242 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 14:
#line 247 "go.y"
    {
		importpkg = gcpkg;
		if(debug['A']) {
			cannedimports("gc.builtin", "package gc\n\n$$\n\n");
		} else {
			cannedimports("gc.builtin", gcimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 15:
#line 259 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 16:
#line 264 "go.y"
    {
		importpkg = profpkg;
		if(debug['A']) {
			cannedimports("prof.builtin", "package prof\n\n$$\n\n");
		} else {
			cannedimports("prof.builtin", profimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 17:
#line 276 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 18:
#line 281 "go.y"
    {
		importpkg = channelspkg;
		if(debug['A']) {
			cannedimports("channels.builtin", "package channels\n\n$$\n\n");
		} else {
			cannedimports("channels.builtin", channelsimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 19:
#line 293 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 20:
#line 298 "go.y"
    {
		importpkg = hashpkg;
		if(debug['A']) {
			cannedimports("hash.builtin", "package hash\n\n$$\n\n");
		} else {
			cannedimports("hash.builtin", hashimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 21:
#line 310 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 22:
#line 315 "go.y"
    {
		importpkg = heapdumppkg;
		if(debug['A']) {
			cannedimports("heapdump.builtin", "package heapdump\n\n$$\n\n");
		} else {
			cannedimports("heapdump.builtin", heapdumpimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 23:
#line 327 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 24:
#line 332 "go.y"
    {
		importpkg = mapspkg;
		if(debug['A']) {
			cannedimports("maps.builtin", "package maps\n\n$$\n\n");
		} else {
			cannedimports("maps.builtin", mapsimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 25:
#line 344 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 26:
#line 349 "go.y"
    {
		importpkg = netpollpkg;
		if(debug['A']) {
			cannedimports("netpoll.builtin", "package netpoll\n\n$$\n\n");
		} else {
			cannedimports("netpoll.builtin", netpollimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 27:
#line 361 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 28:
#line 366 "go.y"
    {
		importpkg = ifacestuffpkg;
		if(debug['A']) {
			cannedimports("ifacestuff.builtin", "package ifacestuff\n\n$$\n\n");
		} else {
			cannedimports("ifacestuff.builtin", ifacestuffimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 29:
#line 378 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 30:
#line 383 "go.y"
    {
		importpkg = vdsopkg;
		if(debug['A']) {
			cannedimports("vdso.builtin", "package vdso\n\n$$\n\n");
		} else {
			cannedimports("vdso.builtin", vdsoimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 31:
#line 395 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 32:
#line 400 "go.y"
    {
		importpkg = printfpkg;
		if(debug['A']) {
			cannedimports("printf.builtin", "package printf\n\n$$\n\n");
		} else {
			cannedimports("printf.builtin", printfimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 33:
#line 412 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 34:
#line 417 "go.y"
    {
		importpkg = stringspkg;
		if(debug['A']) {
			cannedimports("strings.builtin", "package strings\n\n$$\n\n");
		} else {
			cannedimports("strings.builtin", stringsimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 35:
#line 429 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 36:
#line 434 "go.y"
    {
		importpkg = fppkg;
		if(debug['A']) {
			cannedimports("fp.builtin", "package fp\n\n$$\n\n");
		} else {
			cannedimports("fp.builtin", fpimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 37:
#line 446 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 38:
#line 451 "go.y"
    {
		importpkg = schedinitpkg;
		if(debug['A']) {
			cannedimports("schedinit.builtin", "package schedinit\n\n$$\n\n");
		} else {
			cannedimports("schedinit.builtin", schedinitimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 39:
#line 463 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 40:
#line 468 "go.y"
    {
		importpkg = finalizepkg;
		if(debug['A']) {
			cannedimports("finalize.builtin", "package finalize\n\n$$\n\n");
		} else {
			cannedimports("finalize.builtin", finalizeimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 41:
#line 480 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 42:
#line 485 "go.y"
    {
		importpkg = cgopkg;
		if(debug['A']) {
			cannedimports("cgo.builtin", "package cgo\n\n$$\n\n");
		} else {
			cannedimports("cgo.builtin", cgoimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 43:
#line 497 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 44:
#line 502 "go.y"
    {
		importpkg = syncpkg;
		if(debug['A']) {
			cannedimports("sync.builtin", "package sync\n\n$$\n\n");
		} else {
			cannedimports("sync.builtin", syncimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 45:
#line 514 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 46:
#line 519 "go.y"
    {
		importpkg = checkpkg;
		if(debug['A']) {
			cannedimports("check.builtin", "package check\n\n$$\n\n");
		} else {
			cannedimports("check.builtin", checkimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 47:
#line 531 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 48:
#line 536 "go.y"
    {
		importpkg = stackwbpkg;
		if(debug['A']) {
			cannedimports("stackwb.builtin", "package stackwb\n\n$$\n\n");
		} else {
			cannedimports("stackwb.builtin", stackwbimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 49:
#line 548 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 50:
#line 553 "go.y"
    {
		importpkg = deferspkg;
		if(debug['A']) {
			cannedimports("defers.builtin", "package defers\n\n$$\n\n");
		} else {
			cannedimports("defers.builtin", defersimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 51:
#line 565 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 52:
#line 570 "go.y"
    {
		importpkg = seqpkg;
		if(debug['A']) {
			cannedimports("seq.builtin", "package seq\n\n$$\n\n");
		} else {
			cannedimports("seq.builtin", seqimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 53:
#line 582 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 54:
#line 587 "go.y"
    {
		importpkg = runtimepkg;
		if(debug['A']) {
			cannedimports("runtime.builtin", "package runtime\n\n$$\n\n");
		} else {
			cannedimports("runtime.builtin", runtimeimport);
		}

		curio.importsafe = 1;
	}
    break;

  case 55:
#line 599 "go.y"
    {
		importpkg = nil;
	}
    break;

  case 61:
#line 614 "go.y"
    {
		Pkg *ipkg;
		Sym *my;
		Node *pack;
		
		ipkg = importpkg;
		my = importmyname;
		importpkg = nil;
		importmyname = S;

		if(my == nil)
			my = lookup(ipkg->name);

		pack = nod(OPACK, N, N);
		pack->sym = my;
		pack->pkg = ipkg;
		pack->lineno = (yyvsp[(1) - (3)].i);

		if(my->name[0] == '.') {
			importdot(ipkg, pack);
			break;
		}
		if(strcmp(my->name, "init") == 0) {
			yyerror("cannot import package as init - init must be a func");
			break;
		}
		if(my->name[0] == '_' && my->name[1] == '\0')
			break;
		if(my->def) {
			lineno = (yyvsp[(1) - (3)].i);
			redeclare(my, "as imported package name");
		}
		my->def = pack;
		my->lastlineno = (yyvsp[(1) - (3)].i);
		my->block = 1;	// at top level
	}
    break;

  case 62:
#line 651 "go.y"
    {
		// When an invalid import path is passed to importfile,
		// it calls yyerror and then sets up a fake import with
		// no package statement. This allows us to test more
		// than one invalid import statement in a single file.
		if(nerrors == 0)
			fatal("phase error in import");
	}
    break;

  case 65:
#line 666 "go.y"
    {
		// import with original name
		(yyval.i) = parserline();
		importmyname = S;
		importfile(&(yyvsp[(1) - (1)].val), (yyval.i));
	}
    break;

  case 66:
#line 673 "go.y"
    {
		// import with given name
		(yyval.i) = parserline();
		importmyname = (yyvsp[(1) - (2)].sym);
		importfile(&(yyvsp[(2) - (2)].val), (yyval.i));
	}
    break;

  case 67:
#line 680 "go.y"
    {
		// import into my name space
		(yyval.i) = parserline();
		importmyname = lookup(".");
		importfile(&(yyvsp[(2) - (2)].val), (yyval.i));
	}
    break;

  case 68:
#line 689 "go.y"
    {
		if(importpkg->name == nil) {
			importpkg->name = (yyvsp[(2) - (4)].sym)->name;
			pkglookup((yyvsp[(2) - (4)].sym)->name, nil)->npkg++;
		} else if(strcmp(importpkg->name, (yyvsp[(2) - (4)].sym)->name) != 0)
			yyerror("conflicting names %s and %s for package \"%Z\"", importpkg->name, (yyvsp[(2) - (4)].sym)->name, importpkg->path);
		importpkg->direct = 1;
		importpkg->safe = curio.importsafe;

		if(safemode && !curio.importsafe)
			yyerror("cannot import unsafe package \"%Z\"", importpkg->path);
	}
    break;

  case 70:
#line 704 "go.y"
    {
		if(strcmp((yyvsp[(1) - (1)].sym)->name, "safe") == 0)
			curio.importsafe = 1;
	}
    break;

  case 71:
#line 710 "go.y"
    {
		defercheckwidth();
	}
    break;

  case 72:
#line 714 "go.y"
    {
		resumecheckwidth();
		unimportfile();
	}
    break;

  case 73:
#line 723 "go.y"
    {
		yyerror("empty top-level declaration");
		(yyval.list) = nil;
	}
    break;

  case 75:
#line 729 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 76:
#line 733 "go.y"
    {
		yyerror("non-declaration statement outside function body");
		(yyval.list) = nil;
	}
    break;

  case 77:
#line 738 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 78:
#line 744 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (2)].list);
	}
    break;

  case 79:
#line 748 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
	}
    break;

  case 80:
#line 752 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 81:
#line 756 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (2)].list);
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 82:
#line 762 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 83:
#line 768 "go.y"
    {
		(yyval.list) = concat((yyvsp[(3) - (7)].list), (yyvsp[(5) - (7)].list));
		iota = -100000;
		lastconst = nil;
	}
    break;

  case 84:
#line 774 "go.y"
    {
		(yyval.list) = nil;
		iota = -100000;
	}
    break;

  case 85:
#line 779 "go.y"
    {
		(yyval.list) = list1((yyvsp[(2) - (2)].node));
	}
    break;

  case 86:
#line 783 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (5)].list);
	}
    break;

  case 87:
#line 787 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 88:
#line 793 "go.y"
    {
		iota = 0;
	}
    break;

  case 89:
#line 799 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node), nil);
	}
    break;

  case 90:
#line 803 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (4)].list), (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].list));
	}
    break;

  case 91:
#line 807 "go.y"
    {
		(yyval.list) = variter((yyvsp[(1) - (3)].list), nil, (yyvsp[(3) - (3)].list));
	}
    break;

  case 92:
#line 813 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (4)].list), (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].list));
	}
    break;

  case 93:
#line 817 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (3)].list), N, (yyvsp[(3) - (3)].list));
	}
    break;

  case 95:
#line 824 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node), nil);
	}
    break;

  case 96:
#line 828 "go.y"
    {
		(yyval.list) = constiter((yyvsp[(1) - (1)].list), N, nil);
	}
    break;

  case 97:
#line 834 "go.y"
    {
		// different from dclname because the name
		// becomes visible right here, not at the end
		// of the declaration.
		(yyval.node) = typedcl0((yyvsp[(1) - (1)].sym));
	}
    break;

  case 98:
#line 843 "go.y"
    {
		(yyval.node) = typedcl1((yyvsp[(1) - (2)].node), (yyvsp[(2) - (2)].node), 1);
	}
    break;

  case 99:
#line 849 "go.y"
    {
		(yyval.node) = (yyvsp[(1) - (1)].node);

		// These nodes do not carry line numbers.
		// Since a bare name used as an expression is an error,
		// introduce a wrapper node to give the correct line.
		switch((yyval.node)->op) {
		case ONAME:
		case ONONAME:
		case OTYPE:
		case OPACK:
		case OLITERAL:
			(yyval.node) = nod(OPAREN, (yyval.node), N);
			(yyval.node)->implicit = 1;
			break;
		}
	}
    break;

  case 100:
#line 867 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
		(yyval.node)->etype = (yyvsp[(2) - (3)].i);			// rathole to pass opcode
	}
    break;

  case 101:
#line 872 "go.y"
    {
		if((yyvsp[(1) - (3)].list)->next == nil && (yyvsp[(3) - (3)].list)->next == nil) {
			// simple
			(yyval.node) = nod(OAS, (yyvsp[(1) - (3)].list)->n, (yyvsp[(3) - (3)].list)->n);
			break;
		}
		// multiple
		(yyval.node) = nod(OAS2, N, N);
		(yyval.node)->list = (yyvsp[(1) - (3)].list);
		(yyval.node)->rlist = (yyvsp[(3) - (3)].list);
	}
    break;

  case 102:
#line 884 "go.y"
    {
		if((yyvsp[(3) - (3)].list)->n->op == OTYPESW) {
			(yyval.node) = nod(OTYPESW, N, (yyvsp[(3) - (3)].list)->n->right);
			if((yyvsp[(3) - (3)].list)->next != nil)
				yyerror("expr.(type) must be alone in list");
			if((yyvsp[(1) - (3)].list)->next != nil)
				yyerror("argument count mismatch: %d = %d", count((yyvsp[(1) - (3)].list)), 1);
			else if(((yyvsp[(1) - (3)].list)->n->op != ONAME && (yyvsp[(1) - (3)].list)->n->op != OTYPE && (yyvsp[(1) - (3)].list)->n->op != ONONAME) || isblank((yyvsp[(1) - (3)].list)->n))
				yyerror("invalid variable name %N in type switch", (yyvsp[(1) - (3)].list)->n);
			else
				(yyval.node)->left = dclname((yyvsp[(1) - (3)].list)->n->sym);  // it's a colas, so must not re-use an oldname.
			break;
		}
		(yyval.node) = colas((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list), (yyvsp[(2) - (3)].i));
	}
    break;

  case 103:
#line 900 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (2)].node), nodintconst(1));
		(yyval.node)->implicit = 1;
		(yyval.node)->etype = OADD;
	}
    break;

  case 104:
#line 906 "go.y"
    {
		(yyval.node) = nod(OASOP, (yyvsp[(1) - (2)].node), nodintconst(1));
		(yyval.node)->implicit = 1;
		(yyval.node)->etype = OSUB;
	}
    break;

  case 105:
#line 914 "go.y"
    {
		Node *n, *nn;

		// will be converted to OCASE
		// right will point to next case
		// done in casebody()
		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		(yyval.node)->list = (yyvsp[(2) - (3)].list);
		if(typesw != N && typesw->right != N && (n=typesw->right->left) != N) {
			// type switch - declare variable
			nn = newname(n->sym);
			declare(nn, dclcontext);
			(yyval.node)->nname = nn;

			// keep track of the instances for reporting unused
			nn->defn = typesw->right;
		}
	}
    break;

  case 106:
#line 934 "go.y"
    {
		Node *n;

		// will be converted to OCASE
		// right will point to next case
		// done in casebody()
		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		if((yyvsp[(2) - (5)].list)->next == nil)
			n = nod(OAS, (yyvsp[(2) - (5)].list)->n, (yyvsp[(4) - (5)].node));
		else {
			n = nod(OAS2, N, N);
			n->list = (yyvsp[(2) - (5)].list);
			n->rlist = list1((yyvsp[(4) - (5)].node));
		}
		(yyval.node)->list = list1(n);
	}
    break;

  case 107:
#line 952 "go.y"
    {
		// will be converted to OCASE
		// right will point to next case
		// done in casebody()
		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		(yyval.node)->list = list1(colas((yyvsp[(2) - (5)].list), list1((yyvsp[(4) - (5)].node)), (yyvsp[(3) - (5)].i)));
	}
    break;

  case 108:
#line 961 "go.y"
    {
		Node *n, *nn;

		markdcl();
		(yyval.node) = nod(OXCASE, N, N);
		if(typesw != N && typesw->right != N && (n=typesw->right->left) != N) {
			// type switch - declare variable
			nn = newname(n->sym);
			declare(nn, dclcontext);
			(yyval.node)->nname = nn;

			// keep track of the instances for reporting unused
			nn->defn = typesw->right;
		}
	}
    break;

  case 109:
#line 979 "go.y"
    {
		markdcl();
	}
    break;

  case 110:
#line 983 "go.y"
    {
		if((yyvsp[(3) - (4)].list) == nil)
			(yyval.node) = nod(OEMPTY, N, N);
		else
			(yyval.node) = liststmt((yyvsp[(3) - (4)].list));
		popdcl();
	}
    break;

  case 111:
#line 993 "go.y"
    {
		// If the last token read by the lexer was consumed
		// as part of the case, clear it (parser has cleared yychar).
		// If the last token read by the lexer was the lookahead
		// leave it alone (parser has it cached in yychar).
		// This is so that the stmt_list action doesn't look at
		// the case tokens if the stmt_list is empty.
		yylast = yychar;
		(yyvsp[(1) - (1)].node)->xoffset = block;
	}
    break;

  case 112:
#line 1004 "go.y"
    {
		int last;

		// This is the only place in the language where a statement
		// list is not allowed to drop the final semicolon, because
		// it's the only place where a statement list is not followed 
		// by a closing brace.  Handle the error for pedantry.

		// Find the final token of the statement list.
		// yylast is lookahead; yyprev is last of stmt_list
		last = yyprev;

		if(last > 0 && last != ';' && yychar != '}')
			yyerror("missing statement after label");
		(yyval.node) = (yyvsp[(1) - (3)].node);
		(yyval.node)->nbody = (yyvsp[(3) - (3)].list);
		popdcl();
	}
    break;

  case 113:
#line 1024 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 114:
#line 1028 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].node));
	}
    break;

  case 115:
#line 1034 "go.y"
    {
		markdcl();
	}
    break;

  case 116:
#line 1038 "go.y"
    {
		(yyval.list) = (yyvsp[(3) - (4)].list);
		popdcl();
	}
    break;

  case 117:
#line 1045 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(4) - (4)].node));
		(yyval.node)->list = (yyvsp[(1) - (4)].list);
		(yyval.node)->etype = 0;	// := flag
	}
    break;

  case 118:
#line 1051 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(4) - (4)].node));
		(yyval.node)->list = (yyvsp[(1) - (4)].list);
		(yyval.node)->colas = 1;
		colasdefn((yyvsp[(1) - (4)].list), (yyval.node));
	}
    break;

  case 119:
#line 1058 "go.y"
    {
		(yyval.node) = nod(ORANGE, N, (yyvsp[(2) - (2)].node));
		(yyval.node)->etype = 0; // := flag
	}
    break;

  case 120:
#line 1065 "go.y"
    {
		// init ; test ; incr
		if((yyvsp[(5) - (5)].node) != N && (yyvsp[(5) - (5)].node)->colas != 0)
			yyerror("cannot declare in the for-increment");
		(yyval.node) = nod(OFOR, N, N);
		if((yyvsp[(1) - (5)].node) != N)
			(yyval.node)->ninit = list1((yyvsp[(1) - (5)].node));
		(yyval.node)->ntest = (yyvsp[(3) - (5)].node);
		(yyval.node)->nincr = (yyvsp[(5) - (5)].node);
	}
    break;

  case 121:
#line 1076 "go.y"
    {
		// normal test
		(yyval.node) = nod(OFOR, N, N);
		(yyval.node)->ntest = (yyvsp[(1) - (1)].node);
	}
    break;

  case 123:
#line 1085 "go.y"
    {
		(yyval.node) = (yyvsp[(1) - (2)].node);
		(yyval.node)->nbody = concat((yyval.node)->nbody, (yyvsp[(2) - (2)].list));
	}
    break;

  case 124:
#line 1092 "go.y"
    {
		markdcl();
	}
    break;

  case 125:
#line 1096 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (3)].node);
		popdcl();
	}
    break;

  case 126:
#line 1103 "go.y"
    {
		// test
		(yyval.node) = nod(OIF, N, N);
		(yyval.node)->ntest = (yyvsp[(1) - (1)].node);
	}
    break;

  case 127:
#line 1109 "go.y"
    {
		// init ; test
		(yyval.node) = nod(OIF, N, N);
		if((yyvsp[(1) - (3)].node) != N)
			(yyval.node)->ninit = list1((yyvsp[(1) - (3)].node));
		(yyval.node)->ntest = (yyvsp[(3) - (3)].node);
	}
    break;

  case 128:
#line 1120 "go.y"
    {
		markdcl();
	}
    break;

  case 129:
#line 1124 "go.y"
    {
		if((yyvsp[(3) - (3)].node)->ntest == N)
			yyerror("missing condition in if statement");
	}
    break;

  case 130:
#line 1129 "go.y"
    {
		(yyvsp[(3) - (5)].node)->nbody = (yyvsp[(5) - (5)].list);
	}
    break;

  case 131:
#line 1133 "go.y"
    {
		Node *n;
		NodeList *nn;

		(yyval.node) = (yyvsp[(3) - (8)].node);
		n = (yyvsp[(3) - (8)].node);
		popdcl();
		for(nn = concat((yyvsp[(7) - (8)].list), (yyvsp[(8) - (8)].list)); nn; nn = nn->next) {
			if(nn->n->op == OIF)
				popdcl();
			n->nelse = list1(nn->n);
			n = nn->n;
		}
	}
    break;

  case 132:
#line 1150 "go.y"
    {
		markdcl();
	}
    break;

  case 133:
#line 1154 "go.y"
    {
		if((yyvsp[(4) - (5)].node)->ntest == N)
			yyerror("missing condition in if statement");
		(yyvsp[(4) - (5)].node)->nbody = (yyvsp[(5) - (5)].list);
		(yyval.list) = list1((yyvsp[(4) - (5)].node));
	}
    break;

  case 134:
#line 1162 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 135:
#line 1166 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (2)].list), (yyvsp[(2) - (2)].list));
	}
    break;

  case 136:
#line 1171 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 137:
#line 1175 "go.y"
    {
		NodeList *node;
		
		node = mal(sizeof *node);
		node->n = (yyvsp[(2) - (2)].node);
		node->end = node;
		(yyval.list) = node;
	}
    break;

  case 138:
#line 1186 "go.y"
    {
		markdcl();
	}
    break;

  case 139:
#line 1190 "go.y"
    {
		Node *n;
		n = (yyvsp[(3) - (3)].node)->ntest;
		if(n != N && n->op != OTYPESW)
			n = N;
		typesw = nod(OXXX, typesw, n);
	}
    break;

  case 140:
#line 1198 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (7)].node);
		(yyval.node)->op = OSWITCH;
		(yyval.node)->list = (yyvsp[(6) - (7)].list);
		typesw = typesw->left;
		popdcl();
	}
    break;

  case 141:
#line 1208 "go.y"
    {
		typesw = nod(OXXX, typesw, N);
	}
    break;

  case 142:
#line 1212 "go.y"
    {
		(yyval.node) = nod(OSELECT, N, N);
		(yyval.node)->lineno = typesw->lineno;
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
		typesw = typesw->left;
	}
    break;

  case 144:
#line 1225 "go.y"
    {
		(yyval.node) = nod(OOROR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 145:
#line 1229 "go.y"
    {
		(yyval.node) = nod(OANDAND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 146:
#line 1233 "go.y"
    {
		(yyval.node) = nod(OEQ, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 147:
#line 1237 "go.y"
    {
		(yyval.node) = nod(ONE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 148:
#line 1241 "go.y"
    {
		(yyval.node) = nod(OLT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 149:
#line 1245 "go.y"
    {
		(yyval.node) = nod(OLE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 150:
#line 1249 "go.y"
    {
		(yyval.node) = nod(OGE, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 151:
#line 1253 "go.y"
    {
		(yyval.node) = nod(OGT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 152:
#line 1257 "go.y"
    {
		(yyval.node) = nod(OADD, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 153:
#line 1261 "go.y"
    {
		(yyval.node) = nod(OSUB, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 154:
#line 1265 "go.y"
    {
		(yyval.node) = nod(OOR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 155:
#line 1269 "go.y"
    {
		(yyval.node) = nod(OXOR, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 156:
#line 1273 "go.y"
    {
		(yyval.node) = nod(OMUL, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 157:
#line 1277 "go.y"
    {
		(yyval.node) = nod(ODIV, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 158:
#line 1281 "go.y"
    {
		(yyval.node) = nod(OMOD, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 159:
#line 1285 "go.y"
    {
		(yyval.node) = nod(OAND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 160:
#line 1289 "go.y"
    {
		(yyval.node) = nod(OANDNOT, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 161:
#line 1293 "go.y"
    {
		(yyval.node) = nod(OLSH, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 162:
#line 1297 "go.y"
    {
		(yyval.node) = nod(ORSH, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 163:
#line 1302 "go.y"
    {
		(yyval.node) = nod(OSEND, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 165:
#line 1309 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 166:
#line 1313 "go.y"
    {
		if((yyvsp[(2) - (2)].node)->op == OCOMPLIT) {
			// Special case for &T{...}: turn into (*T){...}.
			(yyval.node) = (yyvsp[(2) - (2)].node);
			(yyval.node)->right = nod(OIND, (yyval.node)->right, N);
			(yyval.node)->right->implicit = 1;
		} else {
			(yyval.node) = nod(OADDR, (yyvsp[(2) - (2)].node), N);
		}
	}
    break;

  case 167:
#line 1324 "go.y"
    {
		(yyval.node) = nod(OPLUS, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 168:
#line 1328 "go.y"
    {
		(yyval.node) = nod(OMINUS, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 169:
#line 1332 "go.y"
    {
		(yyval.node) = nod(ONOT, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 170:
#line 1336 "go.y"
    {
		yyerror("the bitwise complement operator is ^");
		(yyval.node) = nod(OCOM, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 171:
#line 1341 "go.y"
    {
		(yyval.node) = nod(OCOM, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 172:
#line 1345 "go.y"
    {
		(yyval.node) = nod(ORECV, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 173:
#line 1355 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (3)].node), N);
	}
    break;

  case 174:
#line 1359 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (5)].node), N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
	}
    break;

  case 175:
#line 1364 "go.y"
    {
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (6)].node), N);
		(yyval.node)->list = (yyvsp[(3) - (6)].list);
		(yyval.node)->isddd = 1;
	}
    break;

  case 176:
#line 1372 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(1) - (1)].val));
	}
    break;

  case 178:
#line 1377 "go.y"
    {
		if((yyvsp[(1) - (3)].node)->op == OPACK) {
			Sym *s;
			s = restrictlookup((yyvsp[(3) - (3)].sym)->name, (yyvsp[(1) - (3)].node)->pkg);
			(yyvsp[(1) - (3)].node)->used = 1;
			(yyval.node) = oldname(s);
			break;
		}
		(yyval.node) = nod(OXDOT, (yyvsp[(1) - (3)].node), newname((yyvsp[(3) - (3)].sym)));
	}
    break;

  case 179:
#line 1388 "go.y"
    {
		(yyval.node) = nod(ODOTTYPE, (yyvsp[(1) - (5)].node), (yyvsp[(4) - (5)].node));
	}
    break;

  case 180:
#line 1392 "go.y"
    {
		(yyval.node) = nod(OTYPESW, N, (yyvsp[(1) - (5)].node));
	}
    break;

  case 181:
#line 1396 "go.y"
    {
		(yyval.node) = nod(OINDEX, (yyvsp[(1) - (4)].node), (yyvsp[(3) - (4)].node));
	}
    break;

  case 182:
#line 1400 "go.y"
    {
		(yyval.node) = nod(OSLICE, (yyvsp[(1) - (6)].node), nod(OKEY, (yyvsp[(3) - (6)].node), (yyvsp[(5) - (6)].node)));
	}
    break;

  case 183:
#line 1404 "go.y"
    {
		if((yyvsp[(5) - (8)].node) == N)
			yyerror("middle index required in 3-index slice");
		if((yyvsp[(7) - (8)].node) == N)
			yyerror("final index required in 3-index slice");
		(yyval.node) = nod(OSLICE3, (yyvsp[(1) - (8)].node), nod(OKEY, (yyvsp[(3) - (8)].node), nod(OKEY, (yyvsp[(5) - (8)].node), (yyvsp[(7) - (8)].node))));
	}
    break;

  case 185:
#line 1413 "go.y"
    {
		// conversion
		(yyval.node) = nod(OCALL, (yyvsp[(1) - (5)].node), N);
		(yyval.node)->list = list1((yyvsp[(3) - (5)].node));
	}
    break;

  case 186:
#line 1419 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (5)].node);
		(yyval.node)->right = (yyvsp[(1) - (5)].node);
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 187:
#line 1426 "go.y"
    {
		(yyval.node) = (yyvsp[(3) - (5)].node);
		(yyval.node)->right = (yyvsp[(1) - (5)].node);
		(yyval.node)->list = (yyvsp[(4) - (5)].list);
	}
    break;

  case 188:
#line 1432 "go.y"
    {
		yyerror("cannot parenthesize type in composite literal");
		(yyval.node) = (yyvsp[(5) - (7)].node);
		(yyval.node)->right = (yyvsp[(2) - (7)].node);
		(yyval.node)->list = (yyvsp[(6) - (7)].list);
	}
    break;

  case 190:
#line 1441 "go.y"
    {
		// composite expression.
		// make node early so we get the right line number.
		(yyval.node) = nod(OCOMPLIT, N, N);
	}
    break;

  case 191:
#line 1449 "go.y"
    {
		(yyval.node) = nod(OKEY, (yyvsp[(1) - (3)].node), (yyvsp[(3) - (3)].node));
	}
    break;

  case 192:
#line 1455 "go.y"
    {
		// These nodes do not carry line numbers.
		// Since a composite literal commonly spans several lines,
		// the line number on errors may be misleading.
		// Introduce a wrapper node to give the correct line.
		(yyval.node) = (yyvsp[(1) - (1)].node);
		switch((yyval.node)->op) {
		case ONAME:
		case ONONAME:
		case OTYPE:
		case OPACK:
		case OLITERAL:
			(yyval.node) = nod(OPAREN, (yyval.node), N);
			(yyval.node)->implicit = 1;
		}
	}
    break;

  case 193:
#line 1472 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (4)].node);
		(yyval.node)->list = (yyvsp[(3) - (4)].list);
	}
    break;

  case 195:
#line 1480 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (4)].node);
		(yyval.node)->list = (yyvsp[(3) - (4)].list);
	}
    break;

  case 197:
#line 1488 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
		
		// Need to know on lhs of := whether there are ( ).
		// Don't bother with the OPAREN in other cases:
		// it's just a waste of memory and time.
		switch((yyval.node)->op) {
		case ONAME:
		case ONONAME:
		case OPACK:
		case OTYPE:
		case OLITERAL:
		case OTYPESW:
			(yyval.node) = nod(OPAREN, (yyval.node), N);
		}
	}
    break;

  case 201:
#line 1514 "go.y"
    {
		(yyval.i) = LBODY;
	}
    break;

  case 202:
#line 1518 "go.y"
    {
		(yyval.i) = '{';
	}
    break;

  case 203:
#line 1529 "go.y"
    {
		if((yyvsp[(1) - (1)].sym) == S)
			(yyval.node) = N;
		else
			(yyval.node) = newname((yyvsp[(1) - (1)].sym));
	}
    break;

  case 204:
#line 1538 "go.y"
    {
		(yyval.node) = dclname((yyvsp[(1) - (1)].sym));
	}
    break;

  case 205:
#line 1543 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 207:
#line 1550 "go.y"
    {
		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		// during imports, unqualified non-exported identifiers are from builtinpkg
		if(importpkg != nil && !exportname((yyvsp[(1) - (1)].sym)->name))
			(yyval.sym) = pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg);
	}
    break;

  case 209:
#line 1558 "go.y"
    {
		(yyval.sym) = S;
	}
    break;

  case 210:
#line 1564 "go.y"
    {
		Pkg *p;

		if((yyvsp[(2) - (4)].val).u.sval->len == 0)
			p = importpkg;
		else {
			if(isbadimport((yyvsp[(2) - (4)].val).u.sval))
				errorexit();
			p = mkpkg((yyvsp[(2) - (4)].val).u.sval);
		}
		(yyval.sym) = pkglookup((yyvsp[(4) - (4)].sym)->name, p);
	}
    break;

  case 211:
#line 1577 "go.y"
    {
		Pkg *p;

		if((yyvsp[(2) - (4)].val).u.sval->len == 0)
			p = importpkg;
		else {
			if(isbadimport((yyvsp[(2) - (4)].val).u.sval))
				errorexit();
			p = mkpkg((yyvsp[(2) - (4)].val).u.sval);
		}
		(yyval.sym) = pkglookup("?", p);
	}
    break;

  case 212:
#line 1592 "go.y"
    {
		(yyval.node) = oldname((yyvsp[(1) - (1)].sym));
		if((yyval.node)->pack != N)
			(yyval.node)->pack->used = 1;
	}
    break;

  case 214:
#line 1612 "go.y"
    {
		yyerror("final argument in variadic function missing type");
		(yyval.node) = nod(ODDD, typenod(typ(TINTER)), N);
	}
    break;

  case 215:
#line 1617 "go.y"
    {
		(yyval.node) = nod(ODDD, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 221:
#line 1628 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
	}
    break;

  case 225:
#line 1637 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 230:
#line 1647 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
	}
    break;

  case 240:
#line 1668 "go.y"
    {
		if((yyvsp[(1) - (3)].node)->op == OPACK) {
			Sym *s;
			s = restrictlookup((yyvsp[(3) - (3)].sym)->name, (yyvsp[(1) - (3)].node)->pkg);
			(yyvsp[(1) - (3)].node)->used = 1;
			(yyval.node) = oldname(s);
			break;
		}
		(yyval.node) = nod(OXDOT, (yyvsp[(1) - (3)].node), newname((yyvsp[(3) - (3)].sym)));
	}
    break;

  case 241:
#line 1681 "go.y"
    {
		(yyval.node) = nod(OTARRAY, (yyvsp[(2) - (4)].node), (yyvsp[(4) - (4)].node));
	}
    break;

  case 242:
#line 1685 "go.y"
    {
		// array literal of nelem
		(yyval.node) = nod(OTARRAY, nod(ODDD, N, N), (yyvsp[(4) - (4)].node));
	}
    break;

  case 243:
#line 1690 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(2) - (2)].node), N);
		(yyval.node)->etype = Cboth;
	}
    break;

  case 244:
#line 1695 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(3) - (3)].node), N);
		(yyval.node)->etype = Csend;
	}
    break;

  case 245:
#line 1700 "go.y"
    {
		(yyval.node) = nod(OTMAP, (yyvsp[(3) - (5)].node), (yyvsp[(5) - (5)].node));
	}
    break;

  case 248:
#line 1708 "go.y"
    {
		(yyval.node) = nod(OIND, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 249:
#line 1714 "go.y"
    {
		(yyval.node) = nod(OTCHAN, (yyvsp[(3) - (3)].node), N);
		(yyval.node)->etype = Crecv;
	}
    break;

  case 250:
#line 1721 "go.y"
    {
		(yyval.node) = nod(OTSTRUCT, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 251:
#line 1727 "go.y"
    {
		(yyval.node) = nod(OTSTRUCT, N, N);
		fixlbrace((yyvsp[(2) - (3)].i));
	}
    break;

  case 252:
#line 1734 "go.y"
    {
		(yyval.node) = nod(OTINTER, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		fixlbrace((yyvsp[(2) - (5)].i));
	}
    break;

  case 253:
#line 1740 "go.y"
    {
		(yyval.node) = nod(OTINTER, N, N);
		fixlbrace((yyvsp[(2) - (3)].i));
	}
    break;

  case 254:
#line 1751 "go.y"
    {
		(yyval.node) = (yyvsp[(2) - (3)].node);
		if((yyval.node) == N)
			break;
		if(noescape && (yyvsp[(3) - (3)].list) != nil)
			yyerror("can only use //go:noescape with external func implementations");
		(yyval.node)->nbody = (yyvsp[(3) - (3)].list);
		(yyval.node)->endlineno = lineno;
		(yyval.node)->noescape = noescape;
		(yyval.node)->nosplit = nosplit;
		(yyval.node)->nowritebarrier = nowritebarrier;
		funcbody((yyval.node));
	}
    break;

  case 255:
#line 1767 "go.y"
    {
		Node *t;

		(yyval.node) = N;
		(yyvsp[(3) - (5)].list) = checkarglist((yyvsp[(3) - (5)].list), 1);

		if(strcmp((yyvsp[(1) - (5)].sym)->name, "init") == 0) {
			(yyvsp[(1) - (5)].sym) = renameinit();
			if((yyvsp[(3) - (5)].list) != nil || (yyvsp[(5) - (5)].list) != nil)
				yyerror("func init must have no arguments and no return values");
		}
		if(strcmp(localpkg->name, "main") == 0 && strcmp((yyvsp[(1) - (5)].sym)->name, "main") == 0) {
			if((yyvsp[(3) - (5)].list) != nil || (yyvsp[(5) - (5)].list) != nil)
				yyerror("func main must have no arguments and no return values");
		}

		t = nod(OTFUNC, N, N);
		t->list = (yyvsp[(3) - (5)].list);
		t->rlist = (yyvsp[(5) - (5)].list);

		(yyval.node) = nod(ODCLFUNC, N, N);
		(yyval.node)->nname = newname((yyvsp[(1) - (5)].sym));
		(yyval.node)->nname->defn = (yyval.node);
		(yyval.node)->nname->ntype = t;		// TODO: check if nname already has an ntype
		declare((yyval.node)->nname, PFUNC);

		funchdr((yyval.node));
	}
    break;

  case 256:
#line 1796 "go.y"
    {
		Node *rcvr, *t;

		(yyval.node) = N;
		(yyvsp[(2) - (8)].list) = checkarglist((yyvsp[(2) - (8)].list), 0);
		(yyvsp[(6) - (8)].list) = checkarglist((yyvsp[(6) - (8)].list), 1);

		if((yyvsp[(2) - (8)].list) == nil) {
			yyerror("method has no receiver");
			break;
		}
		if((yyvsp[(2) - (8)].list)->next != nil) {
			yyerror("method has multiple receivers");
			break;
		}
		rcvr = (yyvsp[(2) - (8)].list)->n;
		if(rcvr->op != ODCLFIELD) {
			yyerror("bad receiver in method");
			break;
		}

		t = nod(OTFUNC, rcvr, N);
		t->list = (yyvsp[(6) - (8)].list);
		t->rlist = (yyvsp[(8) - (8)].list);

		(yyval.node) = nod(ODCLFUNC, N, N);
		(yyval.node)->shortname = newname((yyvsp[(4) - (8)].sym));
		(yyval.node)->nname = methodname1((yyval.node)->shortname, rcvr->right);
		(yyval.node)->nname->defn = (yyval.node);
		(yyval.node)->nname->ntype = t;
		(yyval.node)->nname->nointerface = nointerface;
		declare((yyval.node)->nname, PFUNC);

		funchdr((yyval.node));
	}
    break;

  case 257:
#line 1834 "go.y"
    {
		Sym *s;
		Type *t;

		(yyval.node) = N;

		s = (yyvsp[(1) - (5)].sym);
		t = functype(N, (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list));

		importsym(s, ONAME);
		if(s->def != N && s->def->op == ONAME) {
			if(eqtype(t, s->def->type)) {
				dclcontext = PDISCARD;  // since we skip funchdr below
				break;
			}
			yyerror("inconsistent definition for func %S during import\n\t%T\n\t%T", s, s->def->type, t);
		}

		(yyval.node) = newname(s);
		(yyval.node)->type = t;
		declare((yyval.node), PFUNC);

		funchdr((yyval.node));
	}
    break;

  case 258:
#line 1859 "go.y"
    {
		(yyval.node) = methodname1(newname((yyvsp[(4) - (8)].sym)), (yyvsp[(2) - (8)].list)->n->right); 
		(yyval.node)->type = functype((yyvsp[(2) - (8)].list)->n, (yyvsp[(6) - (8)].list), (yyvsp[(8) - (8)].list));

		checkwidth((yyval.node)->type);
		addmethod((yyvsp[(4) - (8)].sym), (yyval.node)->type, 0, nointerface);
		nointerface = 0;
		funchdr((yyval.node));
		
		// inl.c's inlnode in on a dotmeth node expects to find the inlineable body as
		// (dotmeth's type)->nname->inl, and dotmeth's type has been pulled
		// out by typecheck's lookdot as this $$->ttype.  So by providing
		// this back link here we avoid special casing there.
		(yyval.node)->type->nname = (yyval.node);
	}
    break;

  case 259:
#line 1877 "go.y"
    {
		(yyvsp[(3) - (5)].list) = checkarglist((yyvsp[(3) - (5)].list), 1);
		(yyval.node) = nod(OTFUNC, N, N);
		(yyval.node)->list = (yyvsp[(3) - (5)].list);
		(yyval.node)->rlist = (yyvsp[(5) - (5)].list);
	}
    break;

  case 260:
#line 1885 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 261:
#line 1889 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (3)].list);
		if((yyval.list) == nil)
			(yyval.list) = list1(nod(OEMPTY, N, N));
	}
    break;

  case 262:
#line 1897 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 263:
#line 1901 "go.y"
    {
		(yyval.list) = list1(nod(ODCLFIELD, N, (yyvsp[(1) - (1)].node)));
	}
    break;

  case 264:
#line 1905 "go.y"
    {
		(yyvsp[(2) - (3)].list) = checkarglist((yyvsp[(2) - (3)].list), 0);
		(yyval.list) = (yyvsp[(2) - (3)].list);
	}
    break;

  case 265:
#line 1912 "go.y"
    {
		closurehdr((yyvsp[(1) - (1)].node));
	}
    break;

  case 266:
#line 1918 "go.y"
    {
		(yyval.node) = closurebody((yyvsp[(3) - (4)].list));
		fixlbrace((yyvsp[(2) - (4)].i));
	}
    break;

  case 267:
#line 1923 "go.y"
    {
		(yyval.node) = closurebody(nil);
	}
    break;

  case 268:
#line 1934 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 269:
#line 1938 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(2) - (3)].list));
		if(nsyntaxerrors == 0)
			testdclstack();
		nointerface = 0;
		noescape = 0;
		nosplit = 0;
		nowritebarrier = 0;
	}
    break;

  case 271:
#line 1951 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 273:
#line 1958 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 274:
#line 1964 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 275:
#line 1968 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 277:
#line 1975 "go.y"
    {
		(yyval.list) = concat((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].list));
	}
    break;

  case 278:
#line 1981 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 279:
#line 1985 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 280:
#line 1991 "go.y"
    {
		NodeList *l;

		Node *n;
		l = (yyvsp[(1) - (3)].list);
		if(l == nil) {
			// ? symbol, during import (list1(N) == nil)
			n = (yyvsp[(2) - (3)].node);
			if(n->op == OIND)
				n = n->left;
			n = embedded(n->sym, importpkg);
			n->right = (yyvsp[(2) - (3)].node);
			n->val = (yyvsp[(3) - (3)].val);
			(yyval.list) = list1(n);
			break;
		}

		for(l=(yyvsp[(1) - (3)].list); l; l=l->next) {
			l->n = nod(ODCLFIELD, l->n, (yyvsp[(2) - (3)].node));
			l->n->val = (yyvsp[(3) - (3)].val);
		}
	}
    break;

  case 281:
#line 2014 "go.y"
    {
		(yyvsp[(1) - (2)].node)->val = (yyvsp[(2) - (2)].val);
		(yyval.list) = list1((yyvsp[(1) - (2)].node));
	}
    break;

  case 282:
#line 2019 "go.y"
    {
		(yyvsp[(2) - (4)].node)->val = (yyvsp[(4) - (4)].val);
		(yyval.list) = list1((yyvsp[(2) - (4)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 283:
#line 2025 "go.y"
    {
		(yyvsp[(2) - (3)].node)->right = nod(OIND, (yyvsp[(2) - (3)].node)->right, N);
		(yyvsp[(2) - (3)].node)->val = (yyvsp[(3) - (3)].val);
		(yyval.list) = list1((yyvsp[(2) - (3)].node));
	}
    break;

  case 284:
#line 2031 "go.y"
    {
		(yyvsp[(3) - (5)].node)->right = nod(OIND, (yyvsp[(3) - (5)].node)->right, N);
		(yyvsp[(3) - (5)].node)->val = (yyvsp[(5) - (5)].val);
		(yyval.list) = list1((yyvsp[(3) - (5)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 285:
#line 2038 "go.y"
    {
		(yyvsp[(3) - (5)].node)->right = nod(OIND, (yyvsp[(3) - (5)].node)->right, N);
		(yyvsp[(3) - (5)].node)->val = (yyvsp[(5) - (5)].val);
		(yyval.list) = list1((yyvsp[(3) - (5)].node));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 286:
#line 2047 "go.y"
    {
		Node *n;

		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		n = oldname((yyvsp[(1) - (1)].sym));
		if(n->pack != N)
			n->pack->used = 1;
	}
    break;

  case 287:
#line 2056 "go.y"
    {
		Pkg *pkg;

		if((yyvsp[(1) - (3)].sym)->def == N || (yyvsp[(1) - (3)].sym)->def->op != OPACK) {
			yyerror("%S is not a package", (yyvsp[(1) - (3)].sym));
			pkg = localpkg;
		} else {
			(yyvsp[(1) - (3)].sym)->def->used = 1;
			pkg = (yyvsp[(1) - (3)].sym)->def->pkg;
		}
		(yyval.sym) = restrictlookup((yyvsp[(3) - (3)].sym)->name, pkg);
	}
    break;

  case 288:
#line 2071 "go.y"
    {
		(yyval.node) = embedded((yyvsp[(1) - (1)].sym), localpkg);
	}
    break;

  case 289:
#line 2077 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, (yyvsp[(1) - (2)].node), (yyvsp[(2) - (2)].node));
		ifacedcl((yyval.node));
	}
    break;

  case 290:
#line 2082 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, oldname((yyvsp[(1) - (1)].sym)));
	}
    break;

  case 291:
#line 2086 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, oldname((yyvsp[(2) - (3)].sym)));
		yyerror("cannot parenthesize embedded type");
	}
    break;

  case 292:
#line 2093 "go.y"
    {
		// without func keyword
		(yyvsp[(2) - (4)].list) = checkarglist((yyvsp[(2) - (4)].list), 1);
		(yyval.node) = nod(OTFUNC, fakethis(), N);
		(yyval.node)->list = (yyvsp[(2) - (4)].list);
		(yyval.node)->rlist = (yyvsp[(4) - (4)].list);
	}
    break;

  case 294:
#line 2107 "go.y"
    {
		(yyval.node) = nod(ONONAME, N, N);
		(yyval.node)->sym = (yyvsp[(1) - (2)].sym);
		(yyval.node) = nod(OKEY, (yyval.node), (yyvsp[(2) - (2)].node));
	}
    break;

  case 295:
#line 2113 "go.y"
    {
		(yyval.node) = nod(ONONAME, N, N);
		(yyval.node)->sym = (yyvsp[(1) - (2)].sym);
		(yyval.node) = nod(OKEY, (yyval.node), (yyvsp[(2) - (2)].node));
	}
    break;

  case 297:
#line 2122 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 298:
#line 2126 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 299:
#line 2131 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 300:
#line 2135 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (2)].list);
	}
    break;

  case 301:
#line 2143 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 303:
#line 2148 "go.y"
    {
		(yyval.node) = liststmt((yyvsp[(1) - (1)].list));
	}
    break;

  case 305:
#line 2153 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 311:
#line 2164 "go.y"
    {
		(yyvsp[(1) - (2)].node) = nod(OLABEL, (yyvsp[(1) - (2)].node), N);
		(yyvsp[(1) - (2)].node)->sym = dclstack;  // context, for goto restrictions
	}
    break;

  case 312:
#line 2169 "go.y"
    {
		NodeList *l;

		(yyvsp[(1) - (4)].node)->defn = (yyvsp[(4) - (4)].node);
		l = list1((yyvsp[(1) - (4)].node));
		if((yyvsp[(4) - (4)].node))
			l = list(l, (yyvsp[(4) - (4)].node));
		(yyval.node) = liststmt(l);
	}
    break;

  case 313:
#line 2179 "go.y"
    {
		// will be converted to OFALL
		(yyval.node) = nod(OXFALL, N, N);
		(yyval.node)->xoffset = block;
	}
    break;

  case 314:
#line 2185 "go.y"
    {
		(yyval.node) = nod(OBREAK, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 315:
#line 2189 "go.y"
    {
		(yyval.node) = nod(OCONTINUE, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 316:
#line 2193 "go.y"
    {
		(yyval.node) = nod(OPROC, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 317:
#line 2197 "go.y"
    {
		(yyval.node) = nod(ODEFER, (yyvsp[(2) - (2)].node), N);
	}
    break;

  case 318:
#line 2201 "go.y"
    {
		(yyval.node) = nod(OGOTO, (yyvsp[(2) - (2)].node), N);
		(yyval.node)->sym = dclstack;  // context, for goto restrictions
	}
    break;

  case 319:
#line 2206 "go.y"
    {
		(yyval.node) = nod(ORETURN, N, N);
		(yyval.node)->list = (yyvsp[(2) - (2)].list);
		if((yyval.node)->list == nil && curfn != N) {
			NodeList *l;

			for(l=curfn->dcl; l; l=l->next) {
				if(l->n->class == PPARAM)
					continue;
				if(l->n->class != PPARAMOUT)
					break;
				if(l->n->sym->def != l->n)
					yyerror("%s is shadowed during return", l->n->sym->name);
			}
		}
	}
    break;

  case 320:
#line 2225 "go.y"
    {
		(yyval.list) = nil;
		if((yyvsp[(1) - (1)].node) != N)
			(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 321:
#line 2231 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (3)].list);
		if((yyvsp[(3) - (3)].node) != N)
			(yyval.list) = list((yyval.list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 322:
#line 2239 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 323:
#line 2243 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 324:
#line 2249 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 325:
#line 2253 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 326:
#line 2259 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 327:
#line 2263 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 328:
#line 2269 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 329:
#line 2273 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 330:
#line 2282 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 331:
#line 2286 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 332:
#line 2290 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 333:
#line 2294 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 334:
#line 2299 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 335:
#line 2303 "go.y"
    {
		(yyval.list) = (yyvsp[(1) - (2)].list);
	}
    break;

  case 340:
#line 2317 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 342:
#line 2323 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 344:
#line 2329 "go.y"
    {
		(yyval.node) = N;
	}
    break;

  case 346:
#line 2335 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 348:
#line 2341 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 350:
#line 2347 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 352:
#line 2353 "go.y"
    {
		(yyval.val).ctype = CTxxx;
	}
    break;

  case 354:
#line 2363 "go.y"
    {
		importimport((yyvsp[(2) - (4)].sym), (yyvsp[(3) - (4)].val).u.sval);
	}
    break;

  case 355:
#line 2367 "go.y"
    {
		importvar((yyvsp[(2) - (4)].sym), (yyvsp[(3) - (4)].type));
	}
    break;

  case 356:
#line 2371 "go.y"
    {
		importconst((yyvsp[(2) - (5)].sym), types[TIDEAL], (yyvsp[(4) - (5)].node));
	}
    break;

  case 357:
#line 2375 "go.y"
    {
		importconst((yyvsp[(2) - (6)].sym), (yyvsp[(3) - (6)].type), (yyvsp[(5) - (6)].node));
	}
    break;

  case 358:
#line 2379 "go.y"
    {
		importtype((yyvsp[(2) - (4)].type), (yyvsp[(3) - (4)].type));
	}
    break;

  case 359:
#line 2383 "go.y"
    {
		if((yyvsp[(2) - (4)].node) == N) {
			dclcontext = PEXTERN;  // since we skip the funcbody below
			break;
		}

		(yyvsp[(2) - (4)].node)->inl = (yyvsp[(3) - (4)].list);

		funcbody((yyvsp[(2) - (4)].node));
		importlist = list(importlist, (yyvsp[(2) - (4)].node));

		if(debug['E']) {
			print("import [%Z] func %lN \n", importpkg->path, (yyvsp[(2) - (4)].node));
			if(debug['m'] > 2 && (yyvsp[(2) - (4)].node)->inl)
				print("inl body:%+H\n", (yyvsp[(2) - (4)].node)->inl);
		}
	}
    break;

  case 360:
#line 2403 "go.y"
    {
		(yyval.sym) = (yyvsp[(1) - (1)].sym);
		structpkg = (yyval.sym)->pkg;
	}
    break;

  case 361:
#line 2410 "go.y"
    {
		(yyval.type) = pkgtype((yyvsp[(1) - (1)].sym));
		importsym((yyvsp[(1) - (1)].sym), OTYPE);
	}
    break;

  case 367:
#line 2430 "go.y"
    {
		(yyval.type) = pkgtype((yyvsp[(1) - (1)].sym));
	}
    break;

  case 368:
#line 2434 "go.y"
    {
		// predefined name like uint8
		(yyvsp[(1) - (1)].sym) = pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg);
		if((yyvsp[(1) - (1)].sym)->def == N || (yyvsp[(1) - (1)].sym)->def->op != OTYPE) {
			yyerror("%s is not a type", (yyvsp[(1) - (1)].sym)->name);
			(yyval.type) = T;
		} else
			(yyval.type) = (yyvsp[(1) - (1)].sym)->def->type;
	}
    break;

  case 369:
#line 2444 "go.y"
    {
		(yyval.type) = aindex(N, (yyvsp[(3) - (3)].type));
	}
    break;

  case 370:
#line 2448 "go.y"
    {
		(yyval.type) = aindex(nodlit((yyvsp[(2) - (4)].val)), (yyvsp[(4) - (4)].type));
	}
    break;

  case 371:
#line 2452 "go.y"
    {
		(yyval.type) = maptype((yyvsp[(3) - (5)].type), (yyvsp[(5) - (5)].type));
	}
    break;

  case 372:
#line 2456 "go.y"
    {
		(yyval.type) = tostruct((yyvsp[(3) - (4)].list));
	}
    break;

  case 373:
#line 2460 "go.y"
    {
		(yyval.type) = tointerface((yyvsp[(3) - (4)].list));
	}
    break;

  case 374:
#line 2464 "go.y"
    {
		(yyval.type) = ptrto((yyvsp[(2) - (2)].type));
	}
    break;

  case 375:
#line 2468 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(2) - (2)].type);
		(yyval.type)->chan = Cboth;
	}
    break;

  case 376:
#line 2474 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (4)].type);
		(yyval.type)->chan = Cboth;
	}
    break;

  case 377:
#line 2480 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (3)].type);
		(yyval.type)->chan = Csend;
	}
    break;

  case 378:
#line 2488 "go.y"
    {
		(yyval.type) = typ(TCHAN);
		(yyval.type)->type = (yyvsp[(3) - (3)].type);
		(yyval.type)->chan = Crecv;
	}
    break;

  case 379:
#line 2496 "go.y"
    {
		(yyval.type) = functype(nil, (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list));
	}
    break;

  case 380:
#line 2502 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, typenod((yyvsp[(2) - (3)].type)));
		if((yyvsp[(1) - (3)].sym))
			(yyval.node)->left = newname((yyvsp[(1) - (3)].sym));
		(yyval.node)->val = (yyvsp[(3) - (3)].val);
	}
    break;

  case 381:
#line 2509 "go.y"
    {
		Type *t;
	
		t = typ(TARRAY);
		t->bound = -1;
		t->type = (yyvsp[(3) - (4)].type);

		(yyval.node) = nod(ODCLFIELD, N, typenod(t));
		if((yyvsp[(1) - (4)].sym))
			(yyval.node)->left = newname((yyvsp[(1) - (4)].sym));
		(yyval.node)->isddd = 1;
		(yyval.node)->val = (yyvsp[(4) - (4)].val);
	}
    break;

  case 382:
#line 2525 "go.y"
    {
		Sym *s;
		Pkg *p;

		if((yyvsp[(1) - (3)].sym) != S && strcmp((yyvsp[(1) - (3)].sym)->name, "?") != 0) {
			(yyval.node) = nod(ODCLFIELD, newname((yyvsp[(1) - (3)].sym)), typenod((yyvsp[(2) - (3)].type)));
			(yyval.node)->val = (yyvsp[(3) - (3)].val);
		} else {
			s = (yyvsp[(2) - (3)].type)->sym;
			if(s == S && isptr[(yyvsp[(2) - (3)].type)->etype])
				s = (yyvsp[(2) - (3)].type)->type->sym;
			p = importpkg;
			if((yyvsp[(1) - (3)].sym) != S)
				p = (yyvsp[(1) - (3)].sym)->pkg;
			(yyval.node) = embedded(s, p);
			(yyval.node)->right = typenod((yyvsp[(2) - (3)].type));
			(yyval.node)->val = (yyvsp[(3) - (3)].val);
		}
	}
    break;

  case 383:
#line 2547 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, newname((yyvsp[(1) - (5)].sym)), typenod(functype(fakethis(), (yyvsp[(3) - (5)].list), (yyvsp[(5) - (5)].list))));
	}
    break;

  case 384:
#line 2551 "go.y"
    {
		(yyval.node) = nod(ODCLFIELD, N, typenod((yyvsp[(1) - (1)].type)));
	}
    break;

  case 385:
#line 2556 "go.y"
    {
		(yyval.list) = nil;
	}
    break;

  case 387:
#line 2563 "go.y"
    {
		(yyval.list) = (yyvsp[(2) - (3)].list);
	}
    break;

  case 388:
#line 2567 "go.y"
    {
		(yyval.list) = list1(nod(ODCLFIELD, N, typenod((yyvsp[(1) - (1)].type))));
	}
    break;

  case 389:
#line 2577 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(1) - (1)].val));
	}
    break;

  case 390:
#line 2581 "go.y"
    {
		(yyval.node) = nodlit((yyvsp[(2) - (2)].val));
		switch((yyval.node)->val.ctype){
		case CTINT:
		case CTRUNE:
			mpnegfix((yyval.node)->val.u.xval);
			break;
		case CTFLT:
			mpnegflt((yyval.node)->val.u.fval);
			break;
		case CTCPLX:
			mpnegflt(&(yyval.node)->val.u.cval->real);
			mpnegflt(&(yyval.node)->val.u.cval->imag);
			break;
		default:
			yyerror("bad negated constant");
		}
	}
    break;

  case 391:
#line 2600 "go.y"
    {
		(yyval.node) = oldname(pkglookup((yyvsp[(1) - (1)].sym)->name, builtinpkg));
		if((yyval.node)->op != OLITERAL)
			yyerror("bad constant %S", (yyval.node)->sym);
	}
    break;

  case 393:
#line 2609 "go.y"
    {
		if((yyvsp[(2) - (5)].node)->val.ctype == CTRUNE && (yyvsp[(4) - (5)].node)->val.ctype == CTINT) {
			(yyval.node) = (yyvsp[(2) - (5)].node);
			mpaddfixfix((yyvsp[(2) - (5)].node)->val.u.xval, (yyvsp[(4) - (5)].node)->val.u.xval, 0);
			break;
		}
		(yyvsp[(4) - (5)].node)->val.u.cval->real = (yyvsp[(4) - (5)].node)->val.u.cval->imag;
		mpmovecflt(&(yyvsp[(4) - (5)].node)->val.u.cval->imag, 0.0);
		(yyval.node) = nodcplxlit((yyvsp[(2) - (5)].node)->val, (yyvsp[(4) - (5)].node)->val);
	}
    break;

  case 396:
#line 2625 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 397:
#line 2629 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 398:
#line 2635 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 399:
#line 2639 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;

  case 400:
#line 2645 "go.y"
    {
		(yyval.list) = list1((yyvsp[(1) - (1)].node));
	}
    break;

  case 401:
#line 2649 "go.y"
    {
		(yyval.list) = list((yyvsp[(1) - (3)].list), (yyvsp[(3) - (3)].node));
	}
    break;


/* Line 1267 of yacc.c.  */
#line 5538 "y.tab.c"
      default: break;
    }
  YY_SYMBOL_PRINT ("-> $$ =", yyr1[yyn], &yyval, &yyloc);

  YYPOPSTACK (yylen);
  yylen = 0;
  YY_STACK_PRINT (yyss, yyssp);

  *++yyvsp = yyval;


  /* Now `shift' the result of the reduction.  Determine what state
     that goes to, based on the state we popped back to and the rule
     number reduced by.  */

  yyn = yyr1[yyn];

  yystate = yypgoto[yyn - YYNTOKENS] + *yyssp;
  if (0 <= yystate && yystate <= YYLAST && yycheck[yystate] == *yyssp)
    yystate = yytable[yystate];
  else
    yystate = yydefgoto[yyn - YYNTOKENS];

  goto yynewstate;


/*------------------------------------.
| yyerrlab -- here on detecting error |
`------------------------------------*/
yyerrlab:
  /* If not already recovering from an error, report this error.  */
  if (!yyerrstatus)
    {
      ++yynerrs;
#if ! YYERROR_VERBOSE
      yyerror (YY_("syntax error"));
#else
      {
	YYSIZE_T yysize = yysyntax_error (0, yystate, yychar);
	if (yymsg_alloc < yysize && yymsg_alloc < YYSTACK_ALLOC_MAXIMUM)
	  {
	    YYSIZE_T yyalloc = 2 * yysize;
	    if (! (yysize <= yyalloc && yyalloc <= YYSTACK_ALLOC_MAXIMUM))
	      yyalloc = YYSTACK_ALLOC_MAXIMUM;
	    if (yymsg != yymsgbuf)
	      YYSTACK_FREE (yymsg);
	    yymsg = (char *) YYSTACK_ALLOC (yyalloc);
	    if (yymsg)
	      yymsg_alloc = yyalloc;
	    else
	      {
		yymsg = yymsgbuf;
		yymsg_alloc = sizeof yymsgbuf;
	      }
	  }

	if (0 < yysize && yysize <= yymsg_alloc)
	  {
	    (void) yysyntax_error (yymsg, yystate, yychar);
	    yyerror (yymsg);
	  }
	else
	  {
	    yyerror (YY_("syntax error"));
	    if (yysize != 0)
	      goto yyexhaustedlab;
	  }
      }
#endif
    }



  if (yyerrstatus == 3)
    {
      /* If just tried and failed to reuse look-ahead token after an
	 error, discard it.  */

      if (yychar <= YYEOF)
	{
	  /* Return failure if at end of input.  */
	  if (yychar == YYEOF)
	    YYABORT;
	}
      else
	{
	  yydestruct ("Error: discarding",
		      yytoken, &yylval);
	  yychar = YYEMPTY;
	}
    }

  /* Else will try to reuse look-ahead token after shifting the error
     token.  */
  goto yyerrlab1;


/*---------------------------------------------------.
| yyerrorlab -- error raised explicitly by YYERROR.  |
`---------------------------------------------------*/
yyerrorlab:

  /* Pacify compilers like GCC when the user code never invokes
     YYERROR and the label yyerrorlab therefore never appears in user
     code.  */
  if (/*CONSTCOND*/ 0)
     goto yyerrorlab;

  /* Do not reclaim the symbols of the rule which action triggered
     this YYERROR.  */
  YYPOPSTACK (yylen);
  yylen = 0;
  YY_STACK_PRINT (yyss, yyssp);
  yystate = *yyssp;
  goto yyerrlab1;


/*-------------------------------------------------------------.
| yyerrlab1 -- common code for both syntax error and YYERROR.  |
`-------------------------------------------------------------*/
yyerrlab1:
  yyerrstatus = 3;	/* Each real token shifted decrements this.  */

  for (;;)
    {
      yyn = yypact[yystate];
      if (yyn != YYPACT_NINF)
	{
	  yyn += YYTERROR;
	  if (0 <= yyn && yyn <= YYLAST && yycheck[yyn] == YYTERROR)
	    {
	      yyn = yytable[yyn];
	      if (0 < yyn)
		break;
	    }
	}

      /* Pop the current state because it cannot handle the error token.  */
      if (yyssp == yyss)
	YYABORT;


      yydestruct ("Error: popping",
		  yystos[yystate], yyvsp);
      YYPOPSTACK (1);
      yystate = *yyssp;
      YY_STACK_PRINT (yyss, yyssp);
    }

  if (yyn == YYFINAL)
    YYACCEPT;

  *++yyvsp = yylval;


  /* Shift the error token.  */
  YY_SYMBOL_PRINT ("Shifting", yystos[yyn], yyvsp, yylsp);

  yystate = yyn;
  goto yynewstate;


/*-------------------------------------.
| yyacceptlab -- YYACCEPT comes here.  |
`-------------------------------------*/
yyacceptlab:
  yyresult = 0;
  goto yyreturn;

/*-----------------------------------.
| yyabortlab -- YYABORT comes here.  |
`-----------------------------------*/
yyabortlab:
  yyresult = 1;
  goto yyreturn;

#ifndef yyoverflow
/*-------------------------------------------------.
| yyexhaustedlab -- memory exhaustion comes here.  |
`-------------------------------------------------*/
yyexhaustedlab:
  yyerror (YY_("memory exhausted"));
  yyresult = 2;
  /* Fall through.  */
#endif

yyreturn:
  if (yychar != YYEOF && yychar != YYEMPTY)
     yydestruct ("Cleanup: discarding lookahead",
		 yytoken, &yylval);
  /* Do not reclaim the symbols of the rule which action triggered
     this YYABORT or YYACCEPT.  */
  YYPOPSTACK (yylen);
  YY_STACK_PRINT (yyss, yyssp);
  while (yyssp != yyss)
    {
      yydestruct ("Cleanup: popping",
		  yystos[*yyssp], yyvsp);
      YYPOPSTACK (1);
    }
#ifndef yyoverflow
  if (yyss != yyssa)
    YYSTACK_FREE (yyss);
#endif
#if YYERROR_VERBOSE
  if (yymsg != yymsgbuf)
    YYSTACK_FREE (yymsg);
#endif
  /* Make sure YYID is used.  */
  return YYID (yyresult);
}


#line 2653 "go.y"


static void
fixlbrace(int lbr)
{
	// If the opening brace was an LBODY,
	// set up for another one now that we're done.
	// See comment in lex.c about loophack.
	if(lbr == LBODY)
		loophack = 1;
}


