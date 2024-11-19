!function(){"use strict";var e,t,i=window.location,o=window.document,u=o.getElementById("plausible"),l=u.getAttribute("data-api")||(e=(e=u).src.split("/"),t=e[0],e=e[2],t+"//"+e+"/api/event");function n(e,t){try{if("true"===window.localStorage.plausible_ignore)return n=t,(a="localStorage flag")&&console.warn("Ignoring Event: "+a),void(n&&n.callback&&n.callback())}catch(e){}var n,a={},r=(a.n=e,a.u=i.href,a.d=u.getAttribute("data-domain"),a.r=o.referrer||null,t&&t.meta&&(a.m=JSON.stringify(t.meta)),t&&t.props&&(a.p=t.props),t&&t.revenue&&(a.$=t.revenue),a.h=1,new XMLHttpRequest);r.open("POST",l,!0),r.setRequestHeader("Content-Type","text/plain"),r.send(JSON.stringify(a)),r.onreadystatechange=function(){4===r.readyState&&t&&t.callback&&t.callback({status:r.status})}}var a=window.plausible&&window.plausible.q||[];window.plausible=n;for(var r,s=0;s<a.length;s++)n.apply(this,a[s]);function p(){r=i.pathname,n("pageview")}function c(e){return e&&e.tagName&&"a"===e.tagName.toLowerCase()}window.addEventListener("hashchange",p),"prerender"===o.visibilityState?o.addEventListener("visibilitychange",function(){r||"visible"!==o.visibilityState||p()}):p();var f=1;function v(e){var t,n;if("auxclick"!==e.type||e.button===f)return(t=function(e){for(;e&&(void 0===e.tagName||!c(e)||!e.href);)e=e.parentNode;return e}(e.target))&&t.href&&t.href.split("?")[0],!function e(t,n){if(!t||b<n)return!1;if(h(t))return!0;return e(t.parentNode,n+1)}(t,0)&&(n=t)&&n.href&&n.host&&n.host!==i.host?d(e,t,{name:"Outbound Link: Click",props:{url:t.href}}):void 0}function d(e,t,n){var a,r=!1;function i(){r||(r=!0,window.location=t.href)}!function(e,t){if(!e.defaultPrevented)return t=!t.target||t.target.match(/^_(self|parent|top)$/i),e=!(e.ctrlKey||e.metaKey||e.shiftKey)&&"click"===e.type,t&&e}(e,t)?((a={props:n.props}).revenue=n.revenue,plausible(n.name,a)):((a={props:n.props,callback:i}).revenue=n.revenue,plausible(n.name,a),setTimeout(i,5e3),e.preventDefault())}function m(e){var e=h(e)?e:e&&e.parentNode,t={name:null,props:{},revenue:{}},n=e&&e.classList;if(n)for(var a=0;a<n.length;a++){var r,i,o=n.item(a),u=o.match(/plausible-event-(.+)(=|--)(.+)/),u=(u&&(r=u[1],i=u[3].replace(/\+/g," "),"name"==r.toLowerCase()?t.name=i:t.props[r]=i),o.match(/plausible-revenue-(.+)(=|--)(.+)/));u&&(r=u[1],i=u[3],t.revenue[r]=i)}return t}o.addEventListener("click",v),o.addEventListener("auxclick",v);var b=3;function g(e){if("auxclick"!==e.type||e.button===f){for(var t,n,a,r,i=e.target,o=0;o<=b&&i;o++){if((a=i)&&a.tagName&&"form"===a.tagName.toLowerCase())return;c(i)&&(t=i),h(i)&&(n=i),i=i.parentNode}n&&(r=m(n),t?(r.props.url=t.href,d(e,t,r)):((e={}).props=r.props,e.revenue=r.revenue,plausible(r.name,e)))}}function h(e){var t=e&&e.classList;if(t)for(var n=0;n<t.length;n++)if(t.item(n).match(/plausible-event-name(=|--)(.+)/))return!0;return!1}o.addEventListener("submit",function(e){var t,n=e.target,a=m(n);function r(){t||(t=!0,n.submit())}a.name&&(e.preventDefault(),t=!1,setTimeout(r,5e3),(e={props:a.props,callback:r}).revenue=a.revenue,plausible(a.name,e))}),o.addEventListener("click",g),o.addEventListener("auxclick",g)}();